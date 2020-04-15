/**
 * @fileoverview gRPC-Web generated client stub for kubecarrier.api.v1alpha1
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


import * as grpcWeb from 'grpc-web';

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';

import {
  APIVersion,
  UserInfo,
  VersionRequest} from './kubecarrier_pb';

export class KubecarrierClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: string; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: string; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'text';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodInfoVersion = new grpcWeb.AbstractClientBase.MethodInfo(
    APIVersion,
    (request: VersionRequest) => {
      return request.serializeBinary();
    },
    APIVersion.deserializeBinary
  );

  version(
    request: VersionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: APIVersion) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/kubecarrier.api.v1alpha1.Kubecarrier/Version',
      request,
      metadata || {},
      this.methodInfoVersion,
      callback);
  }

  methodInfoWhoAmI = new grpcWeb.AbstractClientBase.MethodInfo(
    UserInfo,
    (request: google_protobuf_empty_pb.Empty) => {
      return request.serializeBinary();
    },
    UserInfo.deserializeBinary
  );

  whoAmI(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: UserInfo) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/kubecarrier.api.v1alpha1.Kubecarrier/WhoAmI',
      request,
      metadata || {},
      this.methodInfoWhoAmI,
      callback);
  }

}
