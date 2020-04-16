import 'dart:async';

import 'package:kubecarrier/src/generated/kubecarrier.pb.dart';
import 'package:kubecarrier/src/generated/kubecarrier.pbgrpc.dart';

class KubeCarrierExample {
  final KubecarrierClient _service;
  KubeCarrierExample(this._service);

  Future<APIVersion> version() async {
    return await _service.version(VersionRequest());
  }
}
