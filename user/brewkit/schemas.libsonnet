local images = import 'images.libsonnet';

local copy = std.native('copy');

{
    generateGRPC(protoFiles):: {
        local mappedFiles = [copy(protoFile, protoFile) for protoFile in protoFiles],

        from: images.protoc,
        workdir: "/app",
        copy: [
            copy("bin/grpc-generate", "bin/grpc-generate")
        ] + mappedFiles,
        command: 'bin/grpc-generate ' + std.join(' ', protoFiles),
        output: {
            artifact: "/app/api",
            "local": "./api"
        },
    },
}
