// Code generated by thriftgo (0.3.20). DO NOT EDIT.

package http

import (
	"context"
	"fmt"
)

type ReqItem struct {
	Id   *int64  `thrift:"id,1,optional" frugal:"1,optional,i64" json:"MyID"`
	Text *string `thrift:"text,2,optional" frugal:"2,optional,string" json:"text,omitempty"`
}

func NewReqItem() *ReqItem {
	return &ReqItem{}
}

func (p *ReqItem) InitDefault() {
}

var ReqItem_Id_DEFAULT int64

func (p *ReqItem) GetId() (v int64) {
	if !p.IsSetId() {
		return ReqItem_Id_DEFAULT
	}
	return *p.Id
}

var ReqItem_Text_DEFAULT string

func (p *ReqItem) GetText() (v string) {
	if !p.IsSetText() {
		return ReqItem_Text_DEFAULT
	}
	return *p.Text
}
func (p *ReqItem) SetId(val *int64) {
	p.Id = val
}
func (p *ReqItem) SetText(val *string) {
	p.Text = val
}

func (p *ReqItem) IsSetId() bool {
	return p.Id != nil
}

func (p *ReqItem) IsSetText() bool {
	return p.Text != nil
}

func (p *ReqItem) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ReqItem(%+v)", *p)
}

var fieldIDToName_ReqItem = map[int16]string{
	1: "id",
	2: "text",
}

type BizCommonParam struct {
	ApiVersion *int64 `thrift:"api_version,1,optional" frugal:"1,optional,i64" json:"api_version,omitempty"`
	Token      *int32 `thrift:"token,2,optional" frugal:"2,optional,i32" json:"token,omitempty"`
}

func NewBizCommonParam() *BizCommonParam {
	return &BizCommonParam{}
}

func (p *BizCommonParam) InitDefault() {
}

var BizCommonParam_ApiVersion_DEFAULT int64

func (p *BizCommonParam) GetApiVersion() (v int64) {
	if !p.IsSetApiVersion() {
		return BizCommonParam_ApiVersion_DEFAULT
	}
	return *p.ApiVersion
}

var BizCommonParam_Token_DEFAULT int32

func (p *BizCommonParam) GetToken() (v int32) {
	if !p.IsSetToken() {
		return BizCommonParam_Token_DEFAULT
	}
	return *p.Token
}
func (p *BizCommonParam) SetApiVersion(val *int64) {
	p.ApiVersion = val
}
func (p *BizCommonParam) SetToken(val *int32) {
	p.Token = val
}

func (p *BizCommonParam) IsSetApiVersion() bool {
	return p.ApiVersion != nil
}

func (p *BizCommonParam) IsSetToken() bool {
	return p.Token != nil
}

func (p *BizCommonParam) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizCommonParam(%+v)", *p)
}

var fieldIDToName_BizCommonParam = map[int16]string{
	1: "api_version",
	2: "token",
}

type BizRequest struct {
	VInt64      *int64             `thrift:"v_int64,1,optional" frugal:"1,optional,i64" json:"v_int64,omitempty"`
	Text        *string            `thrift:"text,2,optional" frugal:"2,optional,string" json:"text,omitempty"`
	Token       *int32             `thrift:"token,3,optional" frugal:"3,optional,i32" json:"token,omitempty"`
	ReqItemsMap map[int64]*ReqItem `thrift:"req_items_map,4,optional" frugal:"4,optional,map<i64:ReqItem>" json:"req_items_map,omitempty"`
	Some        *ReqItem           `thrift:"some,5,optional" frugal:"5,optional,ReqItem" json:"some,omitempty"`
	ReqItems    []string           `thrift:"req_items,6,optional" frugal:"6,optional,list<string>" json:"req_items,omitempty"`
	ApiVersion  *int32             `thrift:"api_version,7,optional" frugal:"7,optional,i32" json:"api_version,omitempty"`
	Uid         *int64             `thrift:"uid,8,optional" frugal:"8,optional,i64" json:"uid,omitempty"`
	Cids        []int64            `thrift:"cids,9,optional" frugal:"9,optional,list<i64>" json:"cids,omitempty"`
	Vids        []string           `thrift:"vids,10,optional" frugal:"10,optional,list<string>" json:"vids,omitempty"`
}

func NewBizRequest() *BizRequest {
	return &BizRequest{}
}

func (p *BizRequest) InitDefault() {
}

var BizRequest_VInt64_DEFAULT int64

func (p *BizRequest) GetVInt64() (v int64) {
	if !p.IsSetVInt64() {
		return BizRequest_VInt64_DEFAULT
	}
	return *p.VInt64
}

var BizRequest_Text_DEFAULT string

func (p *BizRequest) GetText() (v string) {
	if !p.IsSetText() {
		return BizRequest_Text_DEFAULT
	}
	return *p.Text
}

var BizRequest_Token_DEFAULT int32

func (p *BizRequest) GetToken() (v int32) {
	if !p.IsSetToken() {
		return BizRequest_Token_DEFAULT
	}
	return *p.Token
}

var BizRequest_ReqItemsMap_DEFAULT map[int64]*ReqItem

func (p *BizRequest) GetReqItemsMap() (v map[int64]*ReqItem) {
	if !p.IsSetReqItemsMap() {
		return BizRequest_ReqItemsMap_DEFAULT
	}
	return p.ReqItemsMap
}

var BizRequest_Some_DEFAULT *ReqItem

func (p *BizRequest) GetSome() (v *ReqItem) {
	if !p.IsSetSome() {
		return BizRequest_Some_DEFAULT
	}
	return p.Some
}

var BizRequest_ReqItems_DEFAULT []string

func (p *BizRequest) GetReqItems() (v []string) {
	if !p.IsSetReqItems() {
		return BizRequest_ReqItems_DEFAULT
	}
	return p.ReqItems
}

var BizRequest_ApiVersion_DEFAULT int32

func (p *BizRequest) GetApiVersion() (v int32) {
	if !p.IsSetApiVersion() {
		return BizRequest_ApiVersion_DEFAULT
	}
	return *p.ApiVersion
}

var BizRequest_Uid_DEFAULT int64

func (p *BizRequest) GetUid() (v int64) {
	if !p.IsSetUid() {
		return BizRequest_Uid_DEFAULT
	}
	return *p.Uid
}

var BizRequest_Cids_DEFAULT []int64

func (p *BizRequest) GetCids() (v []int64) {
	if !p.IsSetCids() {
		return BizRequest_Cids_DEFAULT
	}
	return p.Cids
}

var BizRequest_Vids_DEFAULT []string

func (p *BizRequest) GetVids() (v []string) {
	if !p.IsSetVids() {
		return BizRequest_Vids_DEFAULT
	}
	return p.Vids
}
func (p *BizRequest) SetVInt64(val *int64) {
	p.VInt64 = val
}
func (p *BizRequest) SetText(val *string) {
	p.Text = val
}
func (p *BizRequest) SetToken(val *int32) {
	p.Token = val
}
func (p *BizRequest) SetReqItemsMap(val map[int64]*ReqItem) {
	p.ReqItemsMap = val
}
func (p *BizRequest) SetSome(val *ReqItem) {
	p.Some = val
}
func (p *BizRequest) SetReqItems(val []string) {
	p.ReqItems = val
}
func (p *BizRequest) SetApiVersion(val *int32) {
	p.ApiVersion = val
}
func (p *BizRequest) SetUid(val *int64) {
	p.Uid = val
}
func (p *BizRequest) SetCids(val []int64) {
	p.Cids = val
}
func (p *BizRequest) SetVids(val []string) {
	p.Vids = val
}

func (p *BizRequest) IsSetVInt64() bool {
	return p.VInt64 != nil
}

func (p *BizRequest) IsSetText() bool {
	return p.Text != nil
}

func (p *BizRequest) IsSetToken() bool {
	return p.Token != nil
}

func (p *BizRequest) IsSetReqItemsMap() bool {
	return p.ReqItemsMap != nil
}

func (p *BizRequest) IsSetSome() bool {
	return p.Some != nil
}

func (p *BizRequest) IsSetReqItems() bool {
	return p.ReqItems != nil
}

func (p *BizRequest) IsSetApiVersion() bool {
	return p.ApiVersion != nil
}

func (p *BizRequest) IsSetUid() bool {
	return p.Uid != nil
}

func (p *BizRequest) IsSetCids() bool {
	return p.Cids != nil
}

func (p *BizRequest) IsSetVids() bool {
	return p.Vids != nil
}

func (p *BizRequest) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizRequest(%+v)", *p)
}

var fieldIDToName_BizRequest = map[int16]string{
	1:  "v_int64",
	2:  "text",
	3:  "token",
	4:  "req_items_map",
	5:  "some",
	6:  "req_items",
	7:  "api_version",
	8:  "uid",
	9:  "cids",
	10: "vids",
}

type RspItem struct {
	ItemId *int64  `thrift:"item_id,1,optional" frugal:"1,optional,i64" json:"item_id,omitempty"`
	Text   *string `thrift:"text,2,optional" frugal:"2,optional,string" json:"text,omitempty"`
}

func NewRspItem() *RspItem {
	return &RspItem{}
}

func (p *RspItem) InitDefault() {
}

var RspItem_ItemId_DEFAULT int64

func (p *RspItem) GetItemId() (v int64) {
	if !p.IsSetItemId() {
		return RspItem_ItemId_DEFAULT
	}
	return *p.ItemId
}

var RspItem_Text_DEFAULT string

func (p *RspItem) GetText() (v string) {
	if !p.IsSetText() {
		return RspItem_Text_DEFAULT
	}
	return *p.Text
}
func (p *RspItem) SetItemId(val *int64) {
	p.ItemId = val
}
func (p *RspItem) SetText(val *string) {
	p.Text = val
}

func (p *RspItem) IsSetItemId() bool {
	return p.ItemId != nil
}

func (p *RspItem) IsSetText() bool {
	return p.Text != nil
}

func (p *RspItem) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("RspItem(%+v)", *p)
}

var fieldIDToName_RspItem = map[int16]string{
	1: "item_id",
	2: "text",
}

type BizResponse struct {
	T           *string            `thrift:"T,1,optional" frugal:"1,optional,string" json:"T,omitempty"`
	RspItems    map[int64]*RspItem `thrift:"rsp_items,2,optional" frugal:"2,optional,map<i64:RspItem>" json:"rsp_items,omitempty"`
	VEnum       *int32             `thrift:"v_enum,3,optional" frugal:"3,optional,i32" json:"v_enum,omitempty"`
	RspItemList []*RspItem         `thrift:"rsp_item_list,4,optional" frugal:"4,optional,list<RspItem>" json:"rsp_item_list,omitempty"`
	HttpCode    *int32             `thrift:"http_code,5,optional" frugal:"5,optional,i32" json:"http_code,omitempty"`
	ItemCount   []int64            `thrift:"item_count,6,optional" frugal:"6,optional,list<i64>" json:"item_count,omitempty"`
}

func NewBizResponse() *BizResponse {
	return &BizResponse{}
}

func (p *BizResponse) InitDefault() {
}

var BizResponse_T_DEFAULT string

func (p *BizResponse) GetT() (v string) {
	if !p.IsSetT() {
		return BizResponse_T_DEFAULT
	}
	return *p.T
}

var BizResponse_RspItems_DEFAULT map[int64]*RspItem

func (p *BizResponse) GetRspItems() (v map[int64]*RspItem) {
	if !p.IsSetRspItems() {
		return BizResponse_RspItems_DEFAULT
	}
	return p.RspItems
}

var BizResponse_VEnum_DEFAULT int32

func (p *BizResponse) GetVEnum() (v int32) {
	if !p.IsSetVEnum() {
		return BizResponse_VEnum_DEFAULT
	}
	return *p.VEnum
}

var BizResponse_RspItemList_DEFAULT []*RspItem

func (p *BizResponse) GetRspItemList() (v []*RspItem) {
	if !p.IsSetRspItemList() {
		return BizResponse_RspItemList_DEFAULT
	}
	return p.RspItemList
}

var BizResponse_HttpCode_DEFAULT int32

func (p *BizResponse) GetHttpCode() (v int32) {
	if !p.IsSetHttpCode() {
		return BizResponse_HttpCode_DEFAULT
	}
	return *p.HttpCode
}

var BizResponse_ItemCount_DEFAULT []int64

func (p *BizResponse) GetItemCount() (v []int64) {
	if !p.IsSetItemCount() {
		return BizResponse_ItemCount_DEFAULT
	}
	return p.ItemCount
}
func (p *BizResponse) SetT(val *string) {
	p.T = val
}
func (p *BizResponse) SetRspItems(val map[int64]*RspItem) {
	p.RspItems = val
}
func (p *BizResponse) SetVEnum(val *int32) {
	p.VEnum = val
}
func (p *BizResponse) SetRspItemList(val []*RspItem) {
	p.RspItemList = val
}
func (p *BizResponse) SetHttpCode(val *int32) {
	p.HttpCode = val
}
func (p *BizResponse) SetItemCount(val []int64) {
	p.ItemCount = val
}

func (p *BizResponse) IsSetT() bool {
	return p.T != nil
}

func (p *BizResponse) IsSetRspItems() bool {
	return p.RspItems != nil
}

func (p *BizResponse) IsSetVEnum() bool {
	return p.VEnum != nil
}

func (p *BizResponse) IsSetRspItemList() bool {
	return p.RspItemList != nil
}

func (p *BizResponse) IsSetHttpCode() bool {
	return p.HttpCode != nil
}

func (p *BizResponse) IsSetItemCount() bool {
	return p.ItemCount != nil
}

func (p *BizResponse) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizResponse(%+v)", *p)
}

var fieldIDToName_BizResponse = map[int16]string{
	1: "T",
	2: "rsp_items",
	3: "v_enum",
	4: "rsp_item_list",
	5: "http_code",
	6: "item_count",
}

type BizService interface {
	BizMethod1(ctx context.Context, req *BizRequest) (r *BizResponse, err error)

	BizMethod2(ctx context.Context, req *BizRequest) (r *BizResponse, err error)

	BizMethod3(ctx context.Context, req *BizRequest) (r *BizResponse, err error)
}

type BizServiceBizMethod1Args struct {
	Req *BizRequest `thrift:"req,1" frugal:"1,default,BizRequest" json:"req"`
}

func NewBizServiceBizMethod1Args() *BizServiceBizMethod1Args {
	return &BizServiceBizMethod1Args{}
}

func (p *BizServiceBizMethod1Args) InitDefault() {
}

var BizServiceBizMethod1Args_Req_DEFAULT *BizRequest

func (p *BizServiceBizMethod1Args) GetReq() (v *BizRequest) {
	if !p.IsSetReq() {
		return BizServiceBizMethod1Args_Req_DEFAULT
	}
	return p.Req
}
func (p *BizServiceBizMethod1Args) SetReq(val *BizRequest) {
	p.Req = val
}

func (p *BizServiceBizMethod1Args) IsSetReq() bool {
	return p.Req != nil
}

func (p *BizServiceBizMethod1Args) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizServiceBizMethod1Args(%+v)", *p)
}

var fieldIDToName_BizServiceBizMethod1Args = map[int16]string{
	1: "req",
}

type BizServiceBizMethod1Result struct {
	Success *BizResponse `thrift:"success,0,optional" frugal:"0,optional,BizResponse" json:"success,omitempty"`
}

func NewBizServiceBizMethod1Result() *BizServiceBizMethod1Result {
	return &BizServiceBizMethod1Result{}
}

func (p *BizServiceBizMethod1Result) InitDefault() {
}

var BizServiceBizMethod1Result_Success_DEFAULT *BizResponse

func (p *BizServiceBizMethod1Result) GetSuccess() (v *BizResponse) {
	if !p.IsSetSuccess() {
		return BizServiceBizMethod1Result_Success_DEFAULT
	}
	return p.Success
}
func (p *BizServiceBizMethod1Result) SetSuccess(x interface{}) {
	p.Success = x.(*BizResponse)
}

func (p *BizServiceBizMethod1Result) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *BizServiceBizMethod1Result) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizServiceBizMethod1Result(%+v)", *p)
}

var fieldIDToName_BizServiceBizMethod1Result = map[int16]string{
	0: "success",
}

type BizServiceBizMethod2Args struct {
	Req *BizRequest `thrift:"req,1" frugal:"1,default,BizRequest" json:"req"`
}

func NewBizServiceBizMethod2Args() *BizServiceBizMethod2Args {
	return &BizServiceBizMethod2Args{}
}

func (p *BizServiceBizMethod2Args) InitDefault() {
}

var BizServiceBizMethod2Args_Req_DEFAULT *BizRequest

func (p *BizServiceBizMethod2Args) GetReq() (v *BizRequest) {
	if !p.IsSetReq() {
		return BizServiceBizMethod2Args_Req_DEFAULT
	}
	return p.Req
}
func (p *BizServiceBizMethod2Args) SetReq(val *BizRequest) {
	p.Req = val
}

func (p *BizServiceBizMethod2Args) IsSetReq() bool {
	return p.Req != nil
}

func (p *BizServiceBizMethod2Args) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizServiceBizMethod2Args(%+v)", *p)
}

var fieldIDToName_BizServiceBizMethod2Args = map[int16]string{
	1: "req",
}

type BizServiceBizMethod2Result struct {
	Success *BizResponse `thrift:"success,0,optional" frugal:"0,optional,BizResponse" json:"success,omitempty"`
}

func NewBizServiceBizMethod2Result() *BizServiceBizMethod2Result {
	return &BizServiceBizMethod2Result{}
}

func (p *BizServiceBizMethod2Result) InitDefault() {
}

var BizServiceBizMethod2Result_Success_DEFAULT *BizResponse

func (p *BizServiceBizMethod2Result) GetSuccess() (v *BizResponse) {
	if !p.IsSetSuccess() {
		return BizServiceBizMethod2Result_Success_DEFAULT
	}
	return p.Success
}
func (p *BizServiceBizMethod2Result) SetSuccess(x interface{}) {
	p.Success = x.(*BizResponse)
}

func (p *BizServiceBizMethod2Result) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *BizServiceBizMethod2Result) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizServiceBizMethod2Result(%+v)", *p)
}

var fieldIDToName_BizServiceBizMethod2Result = map[int16]string{
	0: "success",
}

type BizServiceBizMethod3Args struct {
	Req *BizRequest `thrift:"req,1" frugal:"1,default,BizRequest" json:"req"`
}

func NewBizServiceBizMethod3Args() *BizServiceBizMethod3Args {
	return &BizServiceBizMethod3Args{}
}

func (p *BizServiceBizMethod3Args) InitDefault() {
}

var BizServiceBizMethod3Args_Req_DEFAULT *BizRequest

func (p *BizServiceBizMethod3Args) GetReq() (v *BizRequest) {
	if !p.IsSetReq() {
		return BizServiceBizMethod3Args_Req_DEFAULT
	}
	return p.Req
}
func (p *BizServiceBizMethod3Args) SetReq(val *BizRequest) {
	p.Req = val
}

func (p *BizServiceBizMethod3Args) IsSetReq() bool {
	return p.Req != nil
}

func (p *BizServiceBizMethod3Args) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizServiceBizMethod3Args(%+v)", *p)
}

var fieldIDToName_BizServiceBizMethod3Args = map[int16]string{
	1: "req",
}

type BizServiceBizMethod3Result struct {
	Success *BizResponse `thrift:"success,0,optional" frugal:"0,optional,BizResponse" json:"success,omitempty"`
}

func NewBizServiceBizMethod3Result() *BizServiceBizMethod3Result {
	return &BizServiceBizMethod3Result{}
}

func (p *BizServiceBizMethod3Result) InitDefault() {
}

var BizServiceBizMethod3Result_Success_DEFAULT *BizResponse

func (p *BizServiceBizMethod3Result) GetSuccess() (v *BizResponse) {
	if !p.IsSetSuccess() {
		return BizServiceBizMethod3Result_Success_DEFAULT
	}
	return p.Success
}
func (p *BizServiceBizMethod3Result) SetSuccess(x interface{}) {
	p.Success = x.(*BizResponse)
}

func (p *BizServiceBizMethod3Result) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *BizServiceBizMethod3Result) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BizServiceBizMethod3Result(%+v)", *p)
}

var fieldIDToName_BizServiceBizMethod3Result = map[int16]string{
	0: "success",
}
