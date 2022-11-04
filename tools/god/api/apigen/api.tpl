syntax = "v1"

info (
    title: //TODO 添加标题
    desc: //TODO 添加描述
    author: "{{.gitUser}}"
    email: "{{.gitEmail}}"
)

type request {
    //TODO 添加成员
}

type response {
    //TODO 添加成员
}

service {{.serviceName}} {
    @handler GetUser //TODO 设置处理器
    get /users/id/:userId(request) returns(response)

    @handler CreateUser //TODO 设置处理器
    post /users/create(request)
}