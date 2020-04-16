///
//  Generated code. Do not modify.
//  source: google/protobuf/type.proto
//
// @dart = 2.3
// ignore_for_file: camel_case_types,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'source_context.pb.dart' as $1;
import 'any.pb.dart' as $2;

import 'type.pbenum.dart';

export 'type.pbenum.dart';

class Type extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('Type',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..aOS(1, 'name')
    ..pc<Field>(2, 'fields', $pb.PbFieldType.PM, subBuilder: Field.create)
    ..pPS(3, 'oneofs')
    ..pc<Option>(4, 'options', $pb.PbFieldType.PM, subBuilder: Option.create)
    ..aOM<$1.SourceContext>(5, 'sourceContext',
        subBuilder: $1.SourceContext.create)
    ..e<Syntax>(6, 'syntax', $pb.PbFieldType.OE,
        defaultOrMaker: Syntax.SYNTAX_PROTO2,
        valueOf: Syntax.valueOf,
        enumValues: Syntax.values)
    ..hasRequiredFields = false;

  Type._() : super();
  factory Type() => create();
  factory Type.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory Type.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  Type clone() => Type()..mergeFromMessage(this);
  Type copyWith(void Function(Type) updates) =>
      super.copyWith((message) => updates(message as Type));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Type create() => Type._();
  Type createEmptyInstance() => create();
  static $pb.PbList<Type> createRepeated() => $pb.PbList<Type>();
  @$core.pragma('dart2js:noInline')
  static Type getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Type>(create);
  static Type _defaultInstance;

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
  $core.List<Field> get fields => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<$core.String> get oneofs => $_getList(2);

  @$pb.TagNumber(4)
  $core.List<Option> get options => $_getList(3);

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
  Syntax get syntax => $_getN(5);
  @$pb.TagNumber(6)
  set syntax(Syntax v) {
    setField(6, v);
  }

  @$pb.TagNumber(6)
  $core.bool hasSyntax() => $_has(5);
  @$pb.TagNumber(6)
  void clearSyntax() => clearField(6);
}

class Field extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('Field',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..e<Field_Kind>(1, 'kind', $pb.PbFieldType.OE,
        defaultOrMaker: Field_Kind.TYPE_UNKNOWN,
        valueOf: Field_Kind.valueOf,
        enumValues: Field_Kind.values)
    ..e<Field_Cardinality>(2, 'cardinality', $pb.PbFieldType.OE,
        defaultOrMaker: Field_Cardinality.CARDINALITY_UNKNOWN,
        valueOf: Field_Cardinality.valueOf,
        enumValues: Field_Cardinality.values)
    ..a<$core.int>(3, 'number', $pb.PbFieldType.O3)
    ..aOS(4, 'name')
    ..aOS(6, 'typeUrl')
    ..a<$core.int>(7, 'oneofIndex', $pb.PbFieldType.O3)
    ..aOB(8, 'packed')
    ..pc<Option>(9, 'options', $pb.PbFieldType.PM, subBuilder: Option.create)
    ..aOS(10, 'jsonName')
    ..aOS(11, 'defaultValue')
    ..hasRequiredFields = false;

  Field._() : super();
  factory Field() => create();
  factory Field.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory Field.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  Field clone() => Field()..mergeFromMessage(this);
  Field copyWith(void Function(Field) updates) =>
      super.copyWith((message) => updates(message as Field));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Field create() => Field._();
  Field createEmptyInstance() => create();
  static $pb.PbList<Field> createRepeated() => $pb.PbList<Field>();
  @$core.pragma('dart2js:noInline')
  static Field getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Field>(create);
  static Field _defaultInstance;

  @$pb.TagNumber(1)
  Field_Kind get kind => $_getN(0);
  @$pb.TagNumber(1)
  set kind(Field_Kind v) {
    setField(1, v);
  }

  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => clearField(1);

  @$pb.TagNumber(2)
  Field_Cardinality get cardinality => $_getN(1);
  @$pb.TagNumber(2)
  set cardinality(Field_Cardinality v) {
    setField(2, v);
  }

  @$pb.TagNumber(2)
  $core.bool hasCardinality() => $_has(1);
  @$pb.TagNumber(2)
  void clearCardinality() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get number => $_getIZ(2);
  @$pb.TagNumber(3)
  set number($core.int v) {
    $_setSignedInt32(2, v);
  }

  @$pb.TagNumber(3)
  $core.bool hasNumber() => $_has(2);
  @$pb.TagNumber(3)
  void clearNumber() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get name => $_getSZ(3);
  @$pb.TagNumber(4)
  set name($core.String v) {
    $_setString(3, v);
  }

  @$pb.TagNumber(4)
  $core.bool hasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearName() => clearField(4);

  @$pb.TagNumber(6)
  $core.String get typeUrl => $_getSZ(4);
  @$pb.TagNumber(6)
  set typeUrl($core.String v) {
    $_setString(4, v);
  }

  @$pb.TagNumber(6)
  $core.bool hasTypeUrl() => $_has(4);
  @$pb.TagNumber(6)
  void clearTypeUrl() => clearField(6);

  @$pb.TagNumber(7)
  $core.int get oneofIndex => $_getIZ(5);
  @$pb.TagNumber(7)
  set oneofIndex($core.int v) {
    $_setSignedInt32(5, v);
  }

  @$pb.TagNumber(7)
  $core.bool hasOneofIndex() => $_has(5);
  @$pb.TagNumber(7)
  void clearOneofIndex() => clearField(7);

  @$pb.TagNumber(8)
  $core.bool get packed => $_getBF(6);
  @$pb.TagNumber(8)
  set packed($core.bool v) {
    $_setBool(6, v);
  }

  @$pb.TagNumber(8)
  $core.bool hasPacked() => $_has(6);
  @$pb.TagNumber(8)
  void clearPacked() => clearField(8);

  @$pb.TagNumber(9)
  $core.List<Option> get options => $_getList(7);

  @$pb.TagNumber(10)
  $core.String get jsonName => $_getSZ(8);
  @$pb.TagNumber(10)
  set jsonName($core.String v) {
    $_setString(8, v);
  }

  @$pb.TagNumber(10)
  $core.bool hasJsonName() => $_has(8);
  @$pb.TagNumber(10)
  void clearJsonName() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get defaultValue => $_getSZ(9);
  @$pb.TagNumber(11)
  set defaultValue($core.String v) {
    $_setString(9, v);
  }

  @$pb.TagNumber(11)
  $core.bool hasDefaultValue() => $_has(9);
  @$pb.TagNumber(11)
  void clearDefaultValue() => clearField(11);
}

class Enum extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('Enum',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..aOS(1, 'name')
    ..pc<EnumValue>(2, 'enumvalue', $pb.PbFieldType.PM,
        subBuilder: EnumValue.create)
    ..pc<Option>(3, 'options', $pb.PbFieldType.PM, subBuilder: Option.create)
    ..aOM<$1.SourceContext>(4, 'sourceContext',
        subBuilder: $1.SourceContext.create)
    ..e<Syntax>(5, 'syntax', $pb.PbFieldType.OE,
        defaultOrMaker: Syntax.SYNTAX_PROTO2,
        valueOf: Syntax.valueOf,
        enumValues: Syntax.values)
    ..hasRequiredFields = false;

  Enum._() : super();
  factory Enum() => create();
  factory Enum.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory Enum.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  Enum clone() => Enum()..mergeFromMessage(this);
  Enum copyWith(void Function(Enum) updates) =>
      super.copyWith((message) => updates(message as Enum));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Enum create() => Enum._();
  Enum createEmptyInstance() => create();
  static $pb.PbList<Enum> createRepeated() => $pb.PbList<Enum>();
  @$core.pragma('dart2js:noInline')
  static Enum getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Enum>(create);
  static Enum _defaultInstance;

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
  $core.List<EnumValue> get enumvalue => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<Option> get options => $_getList(2);

  @$pb.TagNumber(4)
  $1.SourceContext get sourceContext => $_getN(3);
  @$pb.TagNumber(4)
  set sourceContext($1.SourceContext v) {
    setField(4, v);
  }

  @$pb.TagNumber(4)
  $core.bool hasSourceContext() => $_has(3);
  @$pb.TagNumber(4)
  void clearSourceContext() => clearField(4);
  @$pb.TagNumber(4)
  $1.SourceContext ensureSourceContext() => $_ensure(3);

  @$pb.TagNumber(5)
  Syntax get syntax => $_getN(4);
  @$pb.TagNumber(5)
  set syntax(Syntax v) {
    setField(5, v);
  }

  @$pb.TagNumber(5)
  $core.bool hasSyntax() => $_has(4);
  @$pb.TagNumber(5)
  void clearSyntax() => clearField(5);
}

class EnumValue extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('EnumValue',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..aOS(1, 'name')
    ..a<$core.int>(2, 'number', $pb.PbFieldType.O3)
    ..pc<Option>(3, 'options', $pb.PbFieldType.PM, subBuilder: Option.create)
    ..hasRequiredFields = false;

  EnumValue._() : super();
  factory EnumValue() => create();
  factory EnumValue.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory EnumValue.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  EnumValue clone() => EnumValue()..mergeFromMessage(this);
  EnumValue copyWith(void Function(EnumValue) updates) =>
      super.copyWith((message) => updates(message as EnumValue));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static EnumValue create() => EnumValue._();
  EnumValue createEmptyInstance() => create();
  static $pb.PbList<EnumValue> createRepeated() => $pb.PbList<EnumValue>();
  @$core.pragma('dart2js:noInline')
  static EnumValue getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EnumValue>(create);
  static EnumValue _defaultInstance;

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
  $core.int get number => $_getIZ(1);
  @$pb.TagNumber(2)
  set number($core.int v) {
    $_setSignedInt32(1, v);
  }

  @$pb.TagNumber(2)
  $core.bool hasNumber() => $_has(1);
  @$pb.TagNumber(2)
  void clearNumber() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<Option> get options => $_getList(2);
}

class Option extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo('Option',
      package: const $pb.PackageName('google.protobuf'),
      createEmptyInstance: create)
    ..aOS(1, 'name')
    ..aOM<$2.Any>(2, 'value', subBuilder: $2.Any.create)
    ..hasRequiredFields = false;

  Option._() : super();
  factory Option() => create();
  factory Option.fromBuffer($core.List<$core.int> i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(i, r);
  factory Option.fromJson($core.String i,
          [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(i, r);
  Option clone() => Option()..mergeFromMessage(this);
  Option copyWith(void Function(Option) updates) =>
      super.copyWith((message) => updates(message as Option));
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Option create() => Option._();
  Option createEmptyInstance() => create();
  static $pb.PbList<Option> createRepeated() => $pb.PbList<Option>();
  @$core.pragma('dart2js:noInline')
  static Option getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Option>(create);
  static Option _defaultInstance;

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
  $2.Any get value => $_getN(1);
  @$pb.TagNumber(2)
  set value($2.Any v) {
    setField(2, v);
  }

  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => clearField(2);
  @$pb.TagNumber(2)
  $2.Any ensureValue() => $_ensure(1);
}
