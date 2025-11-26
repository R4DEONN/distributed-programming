local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'payment',
];

local proto = [
    'api/client/testinternal/testinternal.proto',
    'api/server/paymentinternal/paymentinternal.proto',
];

project.project(appIDs, proto)