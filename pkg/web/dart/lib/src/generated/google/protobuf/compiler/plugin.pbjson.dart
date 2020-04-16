///
//  Generated code. Do not modify.
//  source: google/protobuf/compiler/plugin.proto
//
// @dart = 2.3
// ignore_for_file: camel_case_types,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type

const Version$json = const {
  '1': 'Version',
  '2': const [
    const {'1': 'major', '3': 1, '4': 1, '5': 5, '10': 'major'},
    const {'1': 'minor', '3': 2, '4': 1, '5': 5, '10': 'minor'},
    const {'1': 'patch', '3': 3, '4': 1, '5': 5, '10': 'patch'},
    const {'1': 'suffix', '3': 4, '4': 1, '5': 9, '10': 'suffix'},
  ],
};

const CodeGeneratorRequest$json = const {
  '1': 'CodeGeneratorRequest',
  '2': const [
    const {
      '1': 'file_to_generate',
      '3': 1,
      '4': 3,
      '5': 9,
      '10': 'fileToGenerate'
    },
    const {'1': 'parameter', '3': 2, '4': 1, '5': 9, '10': 'parameter'},
    const {
      '1': 'proto_file',
      '3': 15,
      '4': 3,
      '5': 11,
      '6': '.google.protobuf.FileDescriptorProto',
      '10': 'protoFile'
    },
    const {
      '1': 'compiler_version',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.compiler.Version',
      '10': 'compilerVersion'
    },
  ],
};

const CodeGeneratorResponse$json = const {
  '1': 'CodeGeneratorResponse',
  '2': const [
    const {'1': 'error', '3': 1, '4': 1, '5': 9, '10': 'error'},
    const {
      '1': 'file',
      '3': 15,
      '4': 3,
      '5': 11,
      '6': '.google.protobuf.compiler.CodeGeneratorResponse.File',
      '10': 'file'
    },
  ],
  '3': const [CodeGeneratorResponse_File$json],
};

const CodeGeneratorResponse_File$json = const {
  '1': 'File',
  '2': const [
    const {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    const {
      '1': 'insertion_point',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'insertionPoint'
    },
    const {'1': 'content', '3': 15, '4': 1, '5': 9, '10': 'content'},
  ],
};
