package retrycall

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/retry"

	"github.com/cloudwego/kitex-tests/pkg/test"
)

func notifyPolicyChange(t *testing.T, jsonStr string, method string) *retry.Container {
	var p retry.Policy
	err := sonic.UnmarshalString(jsonStr, &p)
	test.Assert(t, err == nil, err)
	rc := retry.NewRetryContainer()
	rc.NotifyPolicyChange(method, p)
	return rc
}

// v0.11.0 add type 2 - mixed retry, but the old version has bug that parse the data trigger panic
// json is
// {"enable":true,"type":2,
// "mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}
// }
// actually, the test cases has cover the bug
// this test case just to record the problem, and ignore it in the future
func TestBugfix4InvalidRetry(t *testing.T) {
	mockMethod := "mockMethod"

	// <v0.11.0: NotifyPolicyChange panic
	// >v0.11.0: NotifyPolicyChange successful, mixed retry
	t.Run("type=2, mixed retry", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2,
"mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}
}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Mixed]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["mixed_retry"] != nil, mP["mixed_retry"])
	})

	// expect
	// <v0.11.0: NotifyPolicyChange failed, type not match
	// >v0.11.0: NotifyPolicyChange successful, mixed retry
	t.Run("type=2, mixed retry, but policy has empty failure_policy", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2,
"mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}},
"failure_policy": {}
}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Mixed]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["mixed_retry"] != nil, mP["mixed_retry"])
	})

	// expect
	// <v0.11.0: NotifyPolicyChange failed, panic
	// >v0.11.0: NotifyPolicyChange failed, type not match
	t.Run("type=3, unknown retry with unknown policy", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":3,
"new_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-] failed"), dmap)
		fmt.Println(dmap)
		_, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, !ok)
	})
}

func TestInvalidPolicy4FailureRetry(t *testing.T) {
	mockMethod := "mockMethod"

	t.Run("normal failure retry - success", func(t *testing.T) {
		jsonStr := `{"enable":true,"type":0,
"failure_policy":{"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonStr, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Failure]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["failure_retry"] != nil, mP["failure_retry"])
	})

	t.Run("type is failure retry but policy is empty - success", func(t *testing.T) {
		jsonStr := `{"enable":true,"type":0, "failure_policy":{}}`
		rc := notifyPolicyChange(t, jsonStr, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Failure]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["failure_retry"] != nil, mP["failure_retry"])
	})

	t.Run("type is failure retry but policy is mixed - failed", func(t *testing.T) {
		jsonStr := `{"enable":true,"type":0,
"mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonStr, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Failure] failed"), dmap)
		_, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, !ok)
	})

	t.Run("type is failure retry but no policy - failed", func(t *testing.T) {
		jsonStr := `{"enable":true,"type":0}`
		rc := notifyPolicyChange(t, jsonStr, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Failure] failed"), dmap["msg"])
		_, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, !ok)
	})

	t.Run("type is backup but policy is failure retry - failed", func(t *testing.T) {
		jsonStr := `{"enable":true,"type":1,
"failure_policy":{"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonStr, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "BackupPolicy is nil"), dmap["msg"])
	})

	t.Run("type is failure retry but has multi-policies - success", func(t *testing.T) {
		jsonStr := `{"enable":true,"type":0,
"backup_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}},
"failure_policy":{"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}
}`
		rc := notifyPolicyChange(t, jsonStr, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Failure]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["failure_retry"] != nil, mP["failure_retry"])
	})

	t.Run("type is failure retry but policy json has retry_delay_ms - success", func(t *testing.T) {
		jsonStr := `{"enable":true,"type":0,
"failure_policy":{"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}
}`
		rc := notifyPolicyChange(t, jsonStr, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Failure]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["failure_retry"] != nil, mP["failure_retry"])
	})
}

func TestInvalidPolicy4BackupRequest(t *testing.T) {
	mockMethod := "mockMethod"

	t.Run("normal backup request - success", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":1,
"backup_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Backup]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["backup_request"] != nil, mP)
	})

	t.Run("type is mixedRetry but policy is empty - failed", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2, "backup_policy":{}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "MixedPolicy is nil"), dmap["msg"])
	})

	t.Run("type is backup but no policy - failed", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":1}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "BackupPolicy is nil"), dmap["msg"])
	})

	t.Run("type is failure retry but policy is backupRequest - failed", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":0,
"backup_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "FailurePolicy is nil"), dmap["msg"])
	})

	t.Run("type is backup but has multi-policies - success", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":1,
"backup_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}},
"failure_policy":{"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}
}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Backup]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["backup_request"] != nil, mP)
	})

	t.Run("type is backup but policy json has extra configuration - success", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":1,
"backup_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}
}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Backup]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["backup_request"] != nil, mP)
	})
}

func TestInvalidPolicy4MixedRetry(t *testing.T) {
	mockMethod := "mockMethod"

	t.Run("normal mixed retry - success", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2,
"mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Mixed]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["mixed_retry"] != nil, mP)
	})

	t.Run("mixed retry disable - success but disable", func(t *testing.T) {
		jsonRet := `{"enable":false,"type":2,
"mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "disable retryer[mockMethod-Mixed]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == false, mP)
		test.Assert(t, mP["mixed_retry"] != nil, mP)
	})

	t.Run("type is mixedRetry but policy is empty - failed", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2, "mixed_policy":{}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "invalid retry delay duration in mixedRetryer"), dmap["msg"])
	})

	t.Run("type is mixedRetry but no policy - failed", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "MixedPolicy is nil"), dmap["msg"])
	})

	t.Run("type is backupRequest but policy is mixedRetry - failed", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":1,
"mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "BackupPolicy is nil"), dmap["msg"])
	})

	t.Run("type is mixedRetry but policy is failure retry - failed", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2,
"failure_policy":{"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "MixedPolicy is nil"), dmap["msg"])
	})

	t.Run("type is mixedRetry but has multi-policies - success", func(t *testing.T) {
		jsonRet := `{"enable":true,"type":2,
"mixed_policy":{"retry_delay_ms":10,"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}},
"failure_policy":{"stop_policy":{"max_retry_times":2,"max_duration_ms":0,"disable_chain_stop":false,"ddl_stop":false,"cb_policy":{"error_rate":0.1,"min_sample":200}},"backoff_policy":{"backoff_type":"none"},"retry_same_node":false,"extra":"{\"not_retry_for_timeout\":false,\"rpc_retry_code\":{\"all_error_code\":false,\"error_codes\":[103,1204]},\"biz_retry_code\":{\"all_error_code\":false,\"error_codes\":[]}}","extra_struct":{"not_retry_for_timeout":false,"rpc_retry_code":{"all_error_code":false,"error_codes":[103,1204]},"biz_retry_code":{"all_error_code":false,"error_codes":[]}}}
}`
		rc := notifyPolicyChange(t, jsonRet, mockMethod)

		dmap := rc.Dump().(map[string]interface{})
		test.Assert(t, strings.Contains(dmap["msg"].(string), "new retryer[mockMethod-Mixed]"), dmap["msg"])
		mP, ok := dmap[mockMethod].(map[string]interface{})
		test.Assert(t, ok)
		test.Assert(t, mP["enable"] == true, mP)
		test.Assert(t, mP["mixed_retry"] != nil, mP)
	})
}
