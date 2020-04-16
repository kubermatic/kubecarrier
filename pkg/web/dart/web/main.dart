import 'dart:html';

import 'package:kubecarrier/app.dart';
import 'package:kubecarrier/src/generated/kubecarrier.pbgrpc.dart';
import 'package:grpc/grpc_web.dart';

Future<void> main() async {
  var elem = querySelector('#output');
  elem.text = 'Your Dart app is running....';

  final channel = GrpcWebClientChannel.xhr(Uri.parse('http://localhost:8090'));
  final service = KubecarrierClient(channel);

  final app = KubeCarrierExample(service);
  var version = 'version is ${await app.version()}';
  print(version);
  elem.text = version;
}
