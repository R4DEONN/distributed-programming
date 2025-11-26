local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'product',
];

local proto = [
    'api/client/testinternal/testinternal.proto',
    'api/server/productinternal/productinternal.proto',
];

project.project(appIDs, proto)