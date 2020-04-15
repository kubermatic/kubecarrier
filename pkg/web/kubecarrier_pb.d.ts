import * as jspb from "google-protobuf"

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';

export class APIVersion extends jspb.Message {
  getVersion(): string;
  setVersion(value: string): void;

  getBranch(): string;
  setBranch(value: string): void;

  getBuilddate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setBuilddate(value?: google_protobuf_timestamp_pb.Timestamp): void;
  hasBuilddate(): boolean;
  clearBuilddate(): void;

  getGoversion(): string;
  setGoversion(value: string): void;

  getPlatform(): string;
  setPlatform(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): APIVersion.AsObject;
  static toObject(includeInstance: boolean, msg: APIVersion): APIVersion.AsObject;
  static serializeBinaryToWriter(message: APIVersion, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): APIVersion;
  static deserializeBinaryFromReader(message: APIVersion, reader: jspb.BinaryReader): APIVersion;
}

export namespace APIVersion {
  export type AsObject = {
    version: string,
    branch: string,
    builddate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    goversion: string,
    platform: string,
  }
}

export class VersionRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): VersionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: VersionRequest): VersionRequest.AsObject;
  static serializeBinaryToWriter(message: VersionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): VersionRequest;
  static deserializeBinaryFromReader(message: VersionRequest, reader: jspb.BinaryReader): VersionRequest;
}

export namespace VersionRequest {
  export type AsObject = {
  }
}

export class UserInfo extends jspb.Message {
  getUser(): string;
  setUser(value: string): void;

  getGroupsList(): Array<string>;
  setGroupsList(value: Array<string>): void;
  clearGroupsList(): void;
  addGroups(value: string, index?: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserInfo.AsObject;
  static toObject(includeInstance: boolean, msg: UserInfo): UserInfo.AsObject;
  static serializeBinaryToWriter(message: UserInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserInfo;
  static deserializeBinaryFromReader(message: UserInfo, reader: jspb.BinaryReader): UserInfo;
}

export namespace UserInfo {
  export type AsObject = {
    user: string,
    groupsList: Array<string>,
  }
}
