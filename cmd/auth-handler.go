// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of B33S Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/infobsmi/b33s/internal/auth"
	objectlock "github.com/infobsmi/b33s/internal/bucket/object/lock"
	"github.com/infobsmi/b33s/internal/etag"
	"github.com/infobsmi/b33s/internal/hash"
	xhttp "github.com/infobsmi/b33s/internal/http"
	xjwt "github.com/infobsmi/b33s/internal/jwt"
	"github.com/infobsmi/b33s/internal/logger"
	"github.com/infobsmi/b33s/internal/mcontext"
	"github.com/minio/pkg/bucket/policy"
	iampolicy "github.com/minio/pkg/iam/policy"
)

// Verify if request has JWT.
func isRequestJWT(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get(xhttp.Authorization), jwtAlgorithm)
}

// Verify if request has AWS Signature Version '4'.
func isRequestSignatureV4(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get(xhttp.Authorization), signV4Algorithm)
}

// Verify if request has AWS Signature Version '2'.
func isRequestSignatureV2(r *http.Request) bool {
	return (!strings.HasPrefix(r.Header.Get(xhttp.Authorization), signV4Algorithm) &&
		strings.HasPrefix(r.Header.Get(xhttp.Authorization), signV2Algorithm))
}

// Verify if request has AWS PreSign Version '4'.
func isRequestPresignedSignatureV4(r *http.Request) bool {
	_, ok := r.Form[xhttp.AmzCredential]
	return ok
}

// Verify request has AWS PreSign Version '2'.
func isRequestPresignedSignatureV2(r *http.Request) bool {
	_, ok := r.Form[xhttp.AmzAccessKeyID]
	return ok
}

// Verify if request has AWS Post policy Signature Version '4'.
func isRequestPostPolicySignatureV4(r *http.Request) bool {
	return strings.Contains(r.Header.Get(xhttp.ContentType), "multipart/form-data") &&
		r.Method == http.MethodPost
}

// Verify if the request has AWS Streaming Signature Version '4'. This is only valid for 'PUT' operation.
func isRequestSignStreamingV4(r *http.Request) bool {
	return r.Header.Get(xhttp.AmzContentSha256) == streamingContentSHA256 &&
		r.Method == http.MethodPut
}

// Authorization type.
//
//go:generate stringer -type=authType -trimprefix=authType $GOFILE
type authType int

// List of all supported auth types.
const (
	authTypeUnknown authType = iota
	authTypeAnonymous
	authTypePresigned
	authTypePresignedV2
	authTypePostPolicy
	authTypeStreamingSigned
	authTypeSigned
	authTypeSignedV2
	authTypeJWT
	authTypeSTS
)

// Get request authentication type.
func getRequestAuthType(r *http.Request) authType {
	if r.URL != nil {
		var err error
		r.Form, err = url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			logger.LogIf(r.Context(), err)
			return authTypeUnknown
		}
	}
	if isRequestSignatureV2(r) {
		return authTypeSignedV2
	} else if isRequestPresignedSignatureV2(r) {
		return authTypePresignedV2
	} else if isRequestSignStreamingV4(r) {
		return authTypeStreamingSigned
	} else if isRequestSignatureV4(r) {
		return authTypeSigned
	} else if isRequestPresignedSignatureV4(r) {
		return authTypePresigned
	} else if isRequestJWT(r) {
		return authTypeJWT
	} else if isRequestPostPolicySignatureV4(r) {
		return authTypePostPolicy
	} else if _, ok := r.Form[xhttp.Action]; ok {
		return authTypeSTS
	} else if _, ok := r.Header[xhttp.Authorization]; !ok {
		return authTypeAnonymous
	}
	return authTypeUnknown
}

func validateAdminSignature(ctx context.Context, r *http.Request, region string) (auth.Credentials, map[string]interface{}, bool, APIErrorCode) {
	var cred auth.Credentials
	var owner bool
	s3Err := ErrAccessDenied
	if _, ok := r.Header[xhttp.AmzContentSha256]; ok &&
		getRequestAuthType(r) == authTypeSigned {
		// We only support admin credentials to access admin APIs.
		cred, owner, s3Err = getReqAccessKeyV4(r, region, serviceS3)
		if s3Err != ErrNone {
			return cred, nil, owner, s3Err
		}

		// we only support V4 (no presign) with auth body
		s3Err = isReqAuthenticated(ctx, r, region, serviceS3)
	}
	if s3Err != ErrNone {
		reqInfo := (&logger.ReqInfo{}).AppendTags("requestHeaders", dumpRequest(r))
		ctx := logger.SetReqInfo(ctx, reqInfo)
		logger.LogIf(ctx, errors.New(getAPIError(s3Err).Description), logger.Application)
		return cred, nil, owner, s3Err
	}

	return cred, cred.Claims, owner, ErrNone
}

// checkAdminRequestAuth checks for authentication and authorization for the incoming
// request. It only accepts V2 and V4 requests. Presigned, JWT and anonymous requests
// are automatically rejected.
func checkAdminRequestAuth(ctx context.Context, r *http.Request, action iampolicy.AdminAction, region string) (auth.Credentials, APIErrorCode) {
	cred, claims, owner, s3Err := validateAdminSignature(ctx, r, region)
	if s3Err != ErrNone {
		return cred, s3Err
	}
	if globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     cred.AccessKey,
		Groups:          cred.Groups,
		Action:          iampolicy.Action(action),
		ConditionValues: getConditionValues(r, "", cred.AccessKey, claims),
		IsOwner:         owner,
		Claims:          claims,
	}) {
		// Request is allowed return the appropriate access key.
		return cred, ErrNone
	}

	return cred, ErrAccessDenied
}

// Fetch the security token set by the client.
func getSessionToken(r *http.Request) (token string) {
	token = r.Header.Get(xhttp.AmzSecurityToken)
	if token != "" {
		return token
	}
	return r.Form.Get(xhttp.AmzSecurityToken)
}

// Fetch claims in the security token returned by the client, doesn't return
// errors - upon errors the returned claims map will be empty.
func mustGetClaimsFromToken(r *http.Request) map[string]interface{} {
	claims, _ := getClaimsFromToken(getSessionToken(r))
	return claims
}

func getClaimsFromTokenWithSecret(token, secret string) (map[string]interface{}, error) {
	// JWT token for x-amz-security-token is signed with admin
	// secret key, temporary credentials become invalid if
	// server admin credentials change. This is done to ensure
	// that clients cannot decode the token using the temp
	// secret keys and generate an entirely new claim by essentially
	// hijacking the policies. We need to make sure that this is
	// based an admin credential such that token cannot be decoded
	// on the client side and is treated like an opaque value.
	claims, err := auth.ExtractClaims(token, secret)
	if err != nil {
		if subtle.ConstantTimeCompare([]byte(secret), []byte(globalActiveCred.SecretKey)) == 1 {
			return nil, errAuthentication
		}
		claims, err = auth.ExtractClaims(token, globalActiveCred.SecretKey)
		if err != nil {
			return nil, errAuthentication
		}
	}

	// If AuthZPlugin is set, return without any further checks.
	if newGlobalAuthZPluginFn() != nil {
		return claims.Map(), nil
	}

	// Check if a session policy is set. If so, decode it here.
	sp, spok := claims.Lookup(iampolicy.SessionPolicyName)
	if spok {
		// Looks like subpolicy is set and is a string, if set then its
		// base64 encoded, decode it. Decoding fails reject such
		// requests.
		spBytes, err := base64.StdEncoding.DecodeString(sp)
		if err != nil {
			// Base64 decoding fails, we should log to indicate
			// something is malforming the request sent by client.
			logger.LogIf(GlobalContext, err, logger.Application)
			return nil, errAuthentication
		}
		claims.MapClaims[sessionPolicyNameExtracted] = string(spBytes)
	}

	return claims.Map(), nil
}

// Fetch claims in the security token returned by the client.
func getClaimsFromToken(token string) (map[string]interface{}, error) {
	return getClaimsFromTokenWithSecret(token, globalActiveCred.SecretKey)
}

// Fetch claims in the security token returned by the client and validate the token.
func checkClaimsFromToken(r *http.Request, cred auth.Credentials) (map[string]interface{}, APIErrorCode) {
	token := getSessionToken(r)
	if token != "" && cred.AccessKey == "" {
		// x-amz-security-token is not allowed for anonymous access.
		return nil, ErrNoAccessKey
	}

	if token == "" && cred.IsTemp() {
		// Temporary credentials should always have x-amz-security-token
		return nil, ErrInvalidToken
	}

	if token != "" && !cred.IsTemp() {
		// x-amz-security-token should not present for static credentials.
		return nil, ErrInvalidToken
	}

	if cred.IsTemp() && subtle.ConstantTimeCompare([]byte(token), []byte(cred.SessionToken)) != 1 {
		// validate token for temporary credentials only.
		return nil, ErrInvalidToken
	}

	secret := globalActiveCred.SecretKey
	if cred.IsServiceAccount() {
		token = cred.SessionToken
		secret = cred.SecretKey
	}

	if token != "" {
		claims, err := getClaimsFromTokenWithSecret(token, secret)
		if err != nil {
			return nil, toAPIErrorCode(r.Context(), err)
		}
		return claims, ErrNone
	}

	claims := xjwt.NewMapClaims()
	return claims.Map(), ErrNone
}

// Check request auth type verifies the incoming http request
//   - validates the request signature
//   - validates the policy action if anonymous tests bucket policies if any,
//     for authenticated requests validates IAM policies.
//
// returns APIErrorCode if any to be replied to the client.
func checkRequestAuthType(ctx context.Context, r *http.Request, action policy.Action, bucketName, objectName string) (s3Err APIErrorCode) {
	logger.GetReqInfo(ctx).BucketName = bucketName
	logger.GetReqInfo(ctx).ObjectName = objectName

	_, _, s3Err = checkRequestAuthTypeCredential(ctx, r, action)
	return s3Err
}

func authenticateRequest(ctx context.Context, r *http.Request, action policy.Action) (s3Err APIErrorCode) {
	if logger.GetReqInfo(ctx) == nil {
		logger.LogIf(ctx, errors.New("unexpected context.Context does not have a logger.ReqInfo"), logger.Minio)
		return ErrAccessDenied
	}

	var cred auth.Credentials
	var owner bool
	switch getRequestAuthType(r) {
	case authTypeUnknown, authTypeStreamingSigned:
		return ErrSignatureVersionNotSupported
	case authTypePresignedV2, authTypeSignedV2:
		if s3Err = isReqAuthenticatedV2(r); s3Err != ErrNone {
			return s3Err
		}
		cred, owner, s3Err = getReqAccessKeyV2(r)
	case authTypeSigned, authTypePresigned:
		region := globalSite.Region
		switch action {
		case policy.GetBucketLocationAction, policy.ListAllMyBucketsAction:
			region = ""
		}
		if s3Err = isReqAuthenticated(ctx, r, region, serviceS3); s3Err != ErrNone {
			return s3Err
		}
		cred, owner, s3Err = getReqAccessKeyV4(r, region, serviceS3)
	}
	if s3Err != ErrNone {
		return s3Err
	}

	logger.GetReqInfo(ctx).Cred = cred
	logger.GetReqInfo(ctx).Owner = owner

	// region is valid only for CreateBucketAction.
	var region string
	if action == policy.CreateBucketAction {
		// To extract region from XML in request body, get copy of request body.
		payload, err := io.ReadAll(io.LimitReader(r.Body, maxLocationConstraintSize))
		if err != nil {
			logger.LogIf(ctx, err, logger.Application)
			return ErrMalformedXML
		}

		// Populate payload to extract location constraint.
		r.Body = io.NopCloser(bytes.NewReader(payload))
		region, s3Err = parseLocationConstraint(r)
		if s3Err != ErrNone {
			return s3Err
		}

		// Populate payload again to handle it in HTTP handler.
		r.Body = io.NopCloser(bytes.NewReader(payload))
	}

	logger.GetReqInfo(ctx).Region = region

	return s3Err
}

func authorizeRequest(ctx context.Context, r *http.Request, action policy.Action) (s3Err APIErrorCode) {
	reqInfo := logger.GetReqInfo(ctx)
	if reqInfo == nil {
		return ErrAccessDenied
	}

	cred := reqInfo.Cred
	owner := reqInfo.Owner
	region := reqInfo.Region
	bucket := reqInfo.BucketName
	object := reqInfo.ObjectName

	if action != policy.ListAllMyBucketsAction && cred.AccessKey == "" {
		// Anonymous checks are not meant for ListAllBuckets action
		if globalPolicySys.IsAllowed(policy.Args{
			AccountName:     cred.AccessKey,
			Action:          action,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, region, "", nil),
			IsOwner:         false,
			ObjectName:      object,
		}) {
			// Request is allowed return the appropriate access key.
			return ErrNone
		}

		if action == policy.ListBucketVersionsAction {
			// In AWS S3 s3:ListBucket permission is same as s3:ListBucketVersions permission
			// verify as a fallback.
			if globalPolicySys.IsAllowed(policy.Args{
				AccountName:     cred.AccessKey,
				Action:          policy.ListBucketAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, region, "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				// Request is allowed return the appropriate access key.
				return ErrNone
			}
		}

		return ErrAccessDenied
	}

	if globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     cred.AccessKey,
		Groups:          cred.Groups,
		Action:          iampolicy.Action(action),
		BucketName:      bucket,
		ConditionValues: getConditionValues(r, "", cred.AccessKey, cred.Claims),
		ObjectName:      object,
		IsOwner:         owner,
		Claims:          cred.Claims,
	}) {
		// Request is allowed return the appropriate access key.
		return ErrNone
	}

	if action == policy.ListBucketVersionsAction {
		// In AWS S3 s3:ListBucket permission is same as s3:ListBucketVersions permission
		// verify as a fallback.
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     cred.AccessKey,
			Groups:          cred.Groups,
			Action:          iampolicy.ListBucketAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", cred.AccessKey, cred.Claims),
			ObjectName:      object,
			IsOwner:         owner,
			Claims:          cred.Claims,
		}) {
			// Request is allowed return the appropriate access key.
			return ErrNone
		}
	}

	return ErrAccessDenied
}

// Check request auth type verifies the incoming http request
//   - validates the request signature
//   - validates the policy action if anonymous tests bucket policies if any,
//     for authenticated requests validates IAM policies.
//
// returns APIErrorCode if any to be replied to the client.
// Additionally returns the accessKey used in the request, and if this request is by an admin.
func checkRequestAuthTypeCredential(ctx context.Context, r *http.Request, action policy.Action) (cred auth.Credentials, owner bool, s3Err APIErrorCode) {
	s3Err = authenticateRequest(ctx, r, action)
	reqInfo := logger.GetReqInfo(ctx)
	if reqInfo == nil {
		return cred, owner, ErrAccessDenied
	}

	cred = reqInfo.Cred
	owner = reqInfo.Owner
	if s3Err != ErrNone {
		return cred, owner, s3Err
	}

	return cred, owner, authorizeRequest(ctx, r, action)
}

// Verify if request has valid AWS Signature Version '2'.
func isReqAuthenticatedV2(r *http.Request) (s3Error APIErrorCode) {
	if isRequestSignatureV2(r) {
		return doesSignV2Match(r)
	}
	return doesPresignV2SignatureMatch(r)
}

func reqSignatureV4Verify(r *http.Request, region string, stype serviceType) (s3Error APIErrorCode) {
	sha256sum := getContentSha256Cksum(r, stype)
	switch {
	case isRequestSignatureV4(r):
		return doesSignatureMatch(sha256sum, r, region, stype)
	case isRequestPresignedSignatureV4(r):
		return doesPresignedSignatureMatch(sha256sum, r, region, stype)
	default:
		return ErrAccessDenied
	}
}

// Verify if request has valid AWS Signature Version '4'.
func isReqAuthenticated(ctx context.Context, r *http.Request, region string, stype serviceType) (s3Error APIErrorCode) {
	if errCode := reqSignatureV4Verify(r, region, stype); errCode != ErrNone {
		return errCode
	}

	clientETag, err := etag.FromContentMD5(r.Header)
	if err != nil {
		return ErrInvalidDigest
	}

	// Extract either 'X-Amz-Content-Sha256' header or 'X-Amz-Content-Sha256' query parameter (if V4 presigned)
	// Do not verify 'X-Amz-Content-Sha256' if skipSHA256.
	var contentSHA256 []byte
	if skipSHA256 := skipContentSha256Cksum(r); !skipSHA256 && isRequestPresignedSignatureV4(r) {
		if sha256Sum, ok := r.Form[xhttp.AmzContentSha256]; ok && len(sha256Sum) > 0 {
			contentSHA256, err = hex.DecodeString(sha256Sum[0])
			if err != nil {
				return ErrContentSHA256Mismatch
			}
		}
	} else if _, ok := r.Header[xhttp.AmzContentSha256]; !skipSHA256 && ok {
		contentSHA256, err = hex.DecodeString(r.Header.Get(xhttp.AmzContentSha256))
		if err != nil || len(contentSHA256) == 0 {
			return ErrContentSHA256Mismatch
		}
	}

	// Verify 'Content-Md5' and/or 'X-Amz-Content-Sha256' if present.
	// The verification happens implicit during reading.
	reader, err := hash.NewReader(r.Body, -1, clientETag.String(), hex.EncodeToString(contentSHA256), -1)
	if err != nil {
		return toAPIErrorCode(ctx, err)
	}
	r.Body = reader
	return ErrNone
}

// List of all support S3 auth types.
var supportedS3AuthTypes = map[authType]struct{}{
	authTypeAnonymous:       {},
	authTypePresigned:       {},
	authTypePresignedV2:     {},
	authTypeSigned:          {},
	authTypeSignedV2:        {},
	authTypePostPolicy:      {},
	authTypeStreamingSigned: {},
}

// Validate if the authType is valid and supported.
func isSupportedS3AuthType(aType authType) bool {
	_, ok := supportedS3AuthTypes[aType]
	return ok
}

// setAuthHandler to validate authorization header for the incoming request.
func setAuthHandler(h http.Handler) http.Handler {
	// handler for validating incoming authorization headers.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc, ok := r.Context().Value(mcontext.ContextTraceKey).(*mcontext.TraceCtxt)

		aType := getRequestAuthType(r)
		if aType == authTypeSigned || aType == authTypeSignedV2 || aType == authTypeStreamingSigned {
			// Verify if date headers are set, if not reject the request
			amzDate, errCode := parseAmzDateHeader(r)
			if errCode != ErrNone {
				if ok {
					tc.FuncName = "handler.Auth"
					tc.ResponseRecorder.LogErrBody = true
				}

				// All our internal APIs are sensitive towards Date
				// header, for all requests where Date header is not
				// present we will reject such clients.
				writeErrorResponse(r.Context(), w, errorCodes.ToAPIErr(errCode), r.URL)
				atomic.AddUint64(&globalHTTPStats.rejectedRequestsTime, 1)
				return
			}
			// Verify if the request date header is shifted by less than globalMaxSkewTime parameter in the past
			// or in the future, reject request otherwise.
			curTime := UTCNow()
			if curTime.Sub(amzDate) > globalMaxSkewTime || amzDate.Sub(curTime) > globalMaxSkewTime {
				if ok {
					tc.FuncName = "handler.Auth"
					tc.ResponseRecorder.LogErrBody = true
				}

				writeErrorResponse(r.Context(), w, errorCodes.ToAPIErr(ErrRequestTimeTooSkewed), r.URL)
				atomic.AddUint64(&globalHTTPStats.rejectedRequestsTime, 1)
				return
			}
		}
		if isSupportedS3AuthType(aType) || aType == authTypeJWT || aType == authTypeSTS {
			h.ServeHTTP(w, r)
			return
		}

		if ok {
			tc.FuncName = "handler.Auth"
			tc.ResponseRecorder.LogErrBody = true
		}

		writeErrorResponse(r.Context(), w, errorCodes.ToAPIErr(ErrSignatureVersionNotSupported), r.URL)
		atomic.AddUint64(&globalHTTPStats.rejectedRequestsAuth, 1)
	})
}

func validateSignature(atype authType, r *http.Request) (auth.Credentials, bool, APIErrorCode) {
	var cred auth.Credentials
	var owner bool
	var s3Err APIErrorCode
	switch atype {
	case authTypeUnknown, authTypeStreamingSigned:
		return cred, owner, ErrSignatureVersionNotSupported
	case authTypeSignedV2, authTypePresignedV2:
		if s3Err = isReqAuthenticatedV2(r); s3Err != ErrNone {
			return cred, owner, s3Err
		}
		cred, owner, s3Err = getReqAccessKeyV2(r)
	case authTypePresigned, authTypeSigned:
		region := globalSite.Region
		if s3Err = isReqAuthenticated(GlobalContext, r, region, serviceS3); s3Err != ErrNone {
			return cred, owner, s3Err
		}
		cred, owner, s3Err = getReqAccessKeyV4(r, region, serviceS3)
	}
	if s3Err != ErrNone {
		return cred, owner, s3Err
	}

	return cred, owner, ErrNone
}

func isPutRetentionAllowed(bucketName, objectName string, retDays int, retDate time.Time, retMode objectlock.RetMode, byPassSet bool, r *http.Request, cred auth.Credentials, owner bool) (s3Err APIErrorCode) {
	var retSet bool
	if cred.AccessKey == "" {
		return ErrAccessDenied
	}

	conditions := getConditionValues(r, "", cred.AccessKey, cred.Claims)
	conditions["object-lock-mode"] = []string{string(retMode)}
	conditions["object-lock-retain-until-date"] = []string{retDate.UTC().Format(time.RFC3339)}
	if retDays > 0 {
		conditions["object-lock-remaining-retention-days"] = []string{strconv.Itoa(retDays)}
	}
	if retMode == objectlock.RetGovernance && byPassSet {
		byPassSet = globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     cred.AccessKey,
			Groups:          cred.Groups,
			Action:          iampolicy.BypassGovernanceRetentionAction,
			BucketName:      bucketName,
			ObjectName:      objectName,
			ConditionValues: conditions,
			IsOwner:         owner,
			Claims:          cred.Claims,
		})
	}
	if globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     cred.AccessKey,
		Groups:          cred.Groups,
		Action:          iampolicy.PutObjectRetentionAction,
		BucketName:      bucketName,
		ConditionValues: conditions,
		ObjectName:      objectName,
		IsOwner:         owner,
		Claims:          cred.Claims,
	}) {
		retSet = true
	}
	if byPassSet || retSet {
		return ErrNone
	}
	return ErrAccessDenied
}

// isPutActionAllowed - check if PUT operation is allowed on the resource, this
// call verifies bucket policies and IAM policies, supports multi user
// checks etc.
func isPutActionAllowed(ctx context.Context, atype authType, bucketName, objectName string, r *http.Request, action iampolicy.Action) (s3Err APIErrorCode) {
	var cred auth.Credentials
	var owner bool
	region := globalSite.Region
	switch atype {
	case authTypeUnknown:
		return ErrSignatureVersionNotSupported
	case authTypeSignedV2, authTypePresignedV2:
		cred, owner, s3Err = getReqAccessKeyV2(r)
	case authTypeStreamingSigned, authTypePresigned, authTypeSigned:
		cred, owner, s3Err = getReqAccessKeyV4(r, region, serviceS3)
	}
	if s3Err != ErrNone {
		return s3Err
	}

	logger.GetReqInfo(ctx).Cred = cred
	logger.GetReqInfo(ctx).Owner = owner
	logger.GetReqInfo(ctx).Region = region

	// Do not check for PutObjectRetentionAction permission,
	// if mode and retain until date are not set.
	// Can happen when bucket has default lock config set
	if action == iampolicy.PutObjectRetentionAction &&
		r.Header.Get(xhttp.AmzObjectLockMode) == "" &&
		r.Header.Get(xhttp.AmzObjectLockRetainUntilDate) == "" {
		return ErrNone
	}

	if cred.AccessKey == "" {
		if globalPolicySys.IsAllowed(policy.Args{
			AccountName:     cred.AccessKey,
			Groups:          cred.Groups,
			Action:          policy.Action(action),
			BucketName:      bucketName,
			ConditionValues: getConditionValues(r, "", "", nil),
			IsOwner:         false,
			ObjectName:      objectName,
		}) {
			return ErrNone
		}
		return ErrAccessDenied
	}

	if globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     cred.AccessKey,
		Groups:          cred.Groups,
		Action:          action,
		BucketName:      bucketName,
		ConditionValues: getConditionValues(r, "", cred.AccessKey, cred.Claims),
		ObjectName:      objectName,
		IsOwner:         owner,
		Claims:          cred.Claims,
	}) {
		return ErrNone
	}
	return ErrAccessDenied
}
