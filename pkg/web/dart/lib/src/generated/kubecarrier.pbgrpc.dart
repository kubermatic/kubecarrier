///
//  Generated code. Do not modify.
//  source: kubecarrier.proto
//
// @dart = 2.3
// ignore_for_file: camel_case_types,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'kubecarrier.pb.dart' as $0;
import 'google/protobuf/empty.pb.dart' as $1;
export 'kubecarrier.pb.dart';

class KubecarrierClient extends $grpc.Client {
  static final _$version = $grpc.ClientMethod<$0.VersionRequest, $0.APIVersion>(
      '/kubecarrier.api.v1alpha1.Kubecarrier/Version',
      ($0.VersionRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.APIVersion.fromBuffer(value));
  static final _$whoAmI = $grpc.ClientMethod<$1.Empty, $0.UserInfo>(
      '/kubecarrier.api.v1alpha1.Kubecarrier/WhoAmI',
      ($1.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.UserInfo.fromBuffer(value));
  static final _$versionSteam = $grpc.ClientMethod<$1.Empty, $0.APIVersion>(
      '/kubecarrier.api.v1alpha1.Kubecarrier/VersionSteam',
      ($1.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.APIVersion.fromBuffer(value));

  KubecarrierClient($grpc.ClientChannel channel, {$grpc.CallOptions options})
      : super(channel, options: options);

  $grpc.ResponseFuture<$0.APIVersion> version($0.VersionRequest request,
      {$grpc.CallOptions options}) {
    final call = $createCall(_$version, $async.Stream.fromIterable([request]),
        options: options);
    return $grpc.ResponseFuture(call);
  }

  $grpc.ResponseFuture<$0.UserInfo> whoAmI($1.Empty request,
      {$grpc.CallOptions options}) {
    final call = $createCall(_$whoAmI, $async.Stream.fromIterable([request]),
        options: options);
    return $grpc.ResponseFuture(call);
  }

  $grpc.ResponseStream<$0.APIVersion> versionSteam($1.Empty request,
      {$grpc.CallOptions options}) {
    final call = $createCall(
        _$versionSteam, $async.Stream.fromIterable([request]),
        options: options);
    return $grpc.ResponseStream(call);
  }
}

abstract class KubecarrierServiceBase extends $grpc.Service {
  $core.String get $name => 'kubecarrier.api.v1alpha1.Kubecarrier';

  KubecarrierServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.VersionRequest, $0.APIVersion>(
        'Version',
        version_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.VersionRequest.fromBuffer(value),
        ($0.APIVersion value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $0.UserInfo>(
        'WhoAmI',
        whoAmI_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($0.UserInfo value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $0.APIVersion>(
        'VersionSteam',
        versionSteam_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($0.APIVersion value) => value.writeToBuffer()));
  }

  $async.Future<$0.APIVersion> version_Pre(
      $grpc.ServiceCall call, $async.Future<$0.VersionRequest> request) async {
    return version(call, await request);
  }

  $async.Future<$0.UserInfo> whoAmI_Pre(
      $grpc.ServiceCall call, $async.Future<$1.Empty> request) async {
    return whoAmI(call, await request);
  }

  $async.Stream<$0.APIVersion> versionSteam_Pre(
      $grpc.ServiceCall call, $async.Future<$1.Empty> request) async* {
    yield* versionSteam(call, await request);
  }

  $async.Future<$0.APIVersion> version(
      $grpc.ServiceCall call, $0.VersionRequest request);
  $async.Future<$0.UserInfo> whoAmI($grpc.ServiceCall call, $1.Empty request);
  $async.Stream<$0.APIVersion> versionSteam(
      $grpc.ServiceCall call, $1.Empty request);
}
