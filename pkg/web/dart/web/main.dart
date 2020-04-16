import 'dart:html';

void main() {
  var elem = querySelector('#output');
  elem.text = 'Your Dart app is running....';
  print("I'm super happy to be here with you guys!!!");
  elem.querySelector('#output').text = 'v222';
}
