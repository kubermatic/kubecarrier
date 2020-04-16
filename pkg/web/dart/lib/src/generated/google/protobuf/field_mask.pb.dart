///
//  Generated code. Do not modify.
//  source: google/protobuf/field_mask.proto
//
// @dart = 2.3
// ignore_for_file: camel_case_types,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'package:protobuf/src/protobuf/mixins/well_known.dart' as $mixin;

class FieldMask extends $pb.GeneratedMessage with $mixin.FieldMaskMixin {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('FieldMask',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create,
      toProto3Json: $mixin.FieldMaskMixin.toProto3JsonHelper,
      fromProto3Json: $mixin.FieldMaskMixin.fromProto3JsonHelper)
    ..pPS(1, 'paths')
    ..hasRequiredFields = false;

  FieldMask._() : super();
  factory FieldMask() => create();
  factory FieldMask.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory FieldMask.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  FieldMask clone() => FieldMask()..mergeFromMessage(this);
  FieldMask copyWith(void Function(FieldMask) updates) =>
      super.copyWith((message) => updates(message as FieldMask));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static FieldMask create() => FieldMask._();
  FieldMask createEmptyInstance() => create();
  static $pb.PbList<FieldMask> createRepeated() => $pb.PbList<FieldMask>();
  @$core.pragma('dart2js:noInline')
  static FieldMask getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FieldMask>(create);
  static FieldMask _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get paths => $_getList(0);
}
