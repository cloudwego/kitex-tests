
namespace go fieldmask

struct BizRequest {
    1: string A 
    2: required string B
    3: optional binary RespMask
}

struct BizResponse {
    1: string A 
    2: required string B
    3: string C
}

service BizService {
    BizResponse BizMethod1(1: BizRequest req)
}