///
//  Generated code. Do not modify.
//  source: google/protobuf/api.proto
//
// @dart = 2.3
// ignore_for_file: camel_case_types,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'type.pb.dart' as $3;
import 'source_context.pb.dart' as $1;

import 'type.pbenum.dart' as $3;

class Api extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('Api',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..aOS(1, 'name')
    ..pc<Method>(2, 'methods', $pb.PbFieldType.PM, subBuilder: Method.create)
    ..pc<$3.Option>(3, 'options', $pb.PbFieldType.PM,
        subBuilder: $3.Option.create)
    ..aOS(4, 'version')
    ..aOM<$1.SourceContext>(5, 'sourceContext',
        subBuilder: $1.SourceContext.create)
    ..pc<Mixin>(6, 'mixins', $pb.PbFieldType.PM, subBuilder: Mixin.create)
    ..e<$3.Syntax>(7, 'syntax', $pb.PbFieldType.OE,
        defaultOrMaker: $3.Syntax.SYNTAX_PROTO2,
        valueOf: $3.Syntax.valueOf,
        enumValues: $3.Syntax.values)
    ..hasRequiredFields = false;

  Api._() : super();
  factory Api() => create();
  factory Api.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory Api.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  Api clone() => Api()..mergeFromMessage(this);
  Api copyWith(void Function(Api) updates) =>
      super.copyWith((message) => updates(message as Api));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Api create() => Api._();
  Api createEmptyInstance() => create();
  static $pb.PbList<Api> createRepeated() => $pb.PbList<Api>();
  @$core.pragma('dart2js:noInline')
  static Api getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Api>(create);
  static Api _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String v) {
    $_setString(0, v);
  }

  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<Method> get methods => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<$3.Option> get options => $_getList(2);

  @$pb.TagNumber(4)
  $core.String get version => $_getSZ(3);
  @$pb.TagNumber(4)
  set version($core.String v) {
    $_setString(3, v);
  }

  @$pb.TagNumber(4)
  $core.bool hasVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearVersion() => clearField(4);

  @$pb.TagNumber(5)
  $1.SourceContext get sourceContext => $_getN(4);
  @$pb.TagNumber(5)
  set sourceContext($1.SourceContext v) {
    setField(5, v);
  }

  @$pb.TagNumber(5)
  $core.bool hasSourceContext() => $_has(4);
  @$pb.TagNumber(5)
  void clearSourceContext() => clearField(5);
  @$pb.TagNumber(5)
  $1.SourceContext ensureSourceContext() => $_ensure(4);

  @$pb.TagNumber(6)
  $core.List<Mixin> get mixins => $_getList(5);

  @$pb.TagNumber(7)
  $3.Syntax get syntax => $_getN(6);
  @$pb.TagNumber(7)
  set syntax($3.Syntax v) {
    setField(7, v);
  }

  @$pb.TagNumber(7)
  $core.bool hasSyntax() => $_has(6);
  @$pb.TagNumber(7)
  void clearSyntax() => clearField(7);
}

class Method extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('Method',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..aOS(1, 'name')
    ..aOS(2, 'requestTypeUrl')
    ..aOB(3, 'requestStreaming')
    ..aOS(4, 'responseTypeUrl')
    ..aOB(5, 'responseStreaming')
    ..pc<$3.Option>(6, 'options', $pb.PbFieldType.PM,
        subBuilder: $3.Option.create)
    ..e<$3.Syntax>(7, 'syntax', $pb.PbFieldType.OE,
        defaultOrMaker: $3.Syntax.SYNTAX_PROTO2,
        valueOf: $3.Syntax.valueOf,
        enumValues: $3.Syntax.values)
    ..hasRequiredFields = false;

  Method._() : super();
  factory Method() => create();
  factory Method.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory Method.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  Method clone() => Method()..mergeFromMessage(this);
  Method copyWith(void Function(Method) updates) =>
      super.copyWith((message) => updates(message as Method));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Method create() => Method._();
  Method createEmptyInstance() => create();
  static $pb.PbList<Method> createRepeated() => $pb.PbList<Method>();
  @$core.pragma('dart2js:noInline')
  static Method getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Method>(create);
  static Method _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String v) {
    $_setString(0, v);
  }

  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get requestTypeUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set requestTypeUrl($core.String v) {
    $_setString(1, v);
  }

  @$pb.TagNumber(2)
  $core.bool hasRequestTypeUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearRequestTypeUrl() => clearField(2);

  @$pb.TagNumber(3)
  $core.bool get requestStreaming => $_getBF(2);
  @$pb.TagNumber(3)
  set requestStreaming($core.bool v) {
    $_setBool(2, v);
  }

  @$pb.TagNumber(3)
  $core.bool hasRequestStreaming() => $_has(2);
  @$pb.TagNumber(3)
  void clearRequestStreaming() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get responseTypeUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set responseTypeUrl($core.String v) {
    $_setString(3, v);
  }

  @$pb.TagNumber(4)
  $core.bool hasResponseTypeUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearResponseTypeUrl() => clearField(4);

  @$pb.TagNumber(5)
  $core.bool get responseStreaming => $_getBF(4);
  @$pb.TagNumber(5)
  set responseStreaming($core.bool v) {
    $_setBool(4, v);
  }

  @$pb.TagNumber(5)
  $core.bool hasResponseStreaming() => $_has(4);
  @$pb.TagNumber(5)
  void clearResponseStreaming() => clearField(5);

  @$pb.TagNumber(6)
  $core.List<$3.Option> get options => $_getList(5);

  @$pb.TagNumber(7)
  $3.Syntax get syntax => $_getN(6);
  @$pb.TagNumber(7)
  set syntax($3.Syntax v) {
    setField(7, v);
  }

  @$pb.TagNumber(7)
  $core.bool hasSyntax() => $_has(6);
  @$pb.TagNumber(7)
  void clearSyntax() => clearField(7);
}

class Mixin extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('Mixin',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..aOS(1, 'name')
    ..aOS(2, 'root')
    ..hasRequiredFields = false;

  Mixin._() : super();
  factory Mixin() => create();
  factory Mixin.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory Mixin.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  Mixin clone() => Mixin()..mergeFromMessage(this);
  Mixin copyWith(void Function(Mixin) updates) =>
      super.copyWith((message) => updates(message as Mixin));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Mixin create() => Mixin._();
  Mixin createEmptyInstance() => create();
  static $pb.PbList<Mixin> createRepeated() => $pb.PbList<Mixin>();
  @$core.pragma('dart2js:noInline')
  static Mixin getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Mixin>(create);
  static Mixin _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String v) {
    $_setString(0, v);
  }

  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get root => $_getSZ(1);
  @$pb.TagNumber(2)
  set root($core.String v) {
    $_setString(1, v);
  }

  @$pb.TagNumber(2)
  $core.bool hasRoot() => $_has(1);
  @$pb.TagNumber(2)
  void clearRoot() => clearField(2);
}
