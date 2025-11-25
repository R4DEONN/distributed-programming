local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'userservice',
];

local proto = [
    'api/server/userpublicapi/userpublicapi.proto',
];

project.project(appIDs, proto)