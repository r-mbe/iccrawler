{{plantuml(svg)
@startuml
title EDI对接状态图
skinparam defaultFontName ttf-arphic-uming
skinparam BackgroundColor #fff

state "db" as db {
    state "PO单" as db_PO单
    state "EDI日志" as db_EDI日志
}

state "edi.eos.ickey.cn" as edi_eos {
    state "PO单管理" as edi_eos_PO单管理
}

state "edi.api.ickey.cn" as edi_api {
    state "PO单" as edi_api_PO单
    state "PO单返回" as edi_api_PO单返回
    state "PO单确认" as edi_api_PO单确认
    state "PO变更单" as edi_api_PO变更单
    state "PO变更单确认" as edi_api_PO变更单确认
}

state "openapi.ickey.cn" as openapi {
    state "PO单" as openapi_PO单
    state "PO单返回" as openapi_PO单返回
    state "PO单确认" as openapi_PO单确认
    state "PO变更单" as openapi_PO变更单
    state "PO变更单确认" as openapi_PO变更单确认
}

state "第三方" as 第三方 {
    state "maxim" as maxim
}

state "edi.ickey.cn" as edi {
    state "PO单" as PO单
    state "PO单返回" as PO单返回
    state "PO单确认" as PO单确认
    state "PO变更单" as PO变更单
    state "PO变更单确认" as PO变更单确认
    PO单确认->db:保存回传数据日志
    PO单->db:保存PO请求日志
}

[*]-down->edi_eos_PO单管理
edi_eos_PO单管理->edi_api_PO单:下PO单
edi_api_PO单->openapi_PO单:通过OPENAPI下PO单
edi_api_PO单->PO单:通过EDI下PO单
PO单->第三方:通过EDI下PO单
第三方-up->PO单确认:第三方回传PO单确认消息
PO单确认->edi_api_PO单确认:调用API接口返回确认消息
edi_api_PO单确认->db_erp:更新PO单状态
@enduml
}}