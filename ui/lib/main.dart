import 'package:flutter/material.dart';
import 'dart:math' as math;
import 'dart:async';
import 'package:mqtt_client/mqtt_client.dart';
import 'package:mqtt_client/mqtt_server_client.dart';

void main() {
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(title: Text('MQTT Animated Shape')),
        body: Center(child: AnimatedShape()),
      ),
    );
  }
}

class AnimatedShape extends StatefulWidget {
  @override
  _AnimatedShapeState createState() => _AnimatedShapeState();
}

class _AnimatedShapeState extends State<AnimatedShape>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _rotationAnimation;
  late Animation<double> _radiusAnimation;
  late List<Color> _dotColors;
  late MqttServerClient _client;
  double _mqttRotationFactor = 1.0;
  double _mqttRadiusFactor = 1.0;

  Color _randomColor() {
    return Color((math.Random().nextDouble() * 0xFFFFFF).toInt() << 0)
        .withOpacity(1.0);
  }

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(seconds: 5),
      vsync: this,
    )..repeat();

    _rotationAnimation =
        Tween<double>(begin: 0, end: 2 * math.pi).animate(_controller);
    _radiusAnimation = Tween<double>(begin: 1, end: 0.2).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );

    _dotColors = List.generate(28, (_) => _randomColor());

    Timer.periodic(Duration(milliseconds: 1000), (_) {
      setState(() {
        _dotColors = List.generate(28, (_) => _randomColor());
      });
    });

    _setupMqttClient();
    _connectToServer();
  }

  void _setupMqttClient() {
    _client = MqttServerClient('SERVER', 'flutter_client');
    _client.port = 1883; // Default port for MQTT
    _client.keepAlivePeriod = 20;
    _client.onDisconnected = _onDisconnected;
    _client.onConnected = _onConnected;
    _client.onSubscribed = _onSubscribed;

    final connMess = MqttConnectMessage()
        .authenticateAs('USER', 'PASSWORD')
        .startClean()
        .withWillQos(MqttQos.atLeastOnce);
    _client.connectionMessage = connMess;
  }

  Future<void> _connectToServer() async {
    try {
      await _client.connect();
    } catch (e) {
      print('Exception: $e');
      _client.disconnect();
    }

    if (_client.connectionStatus!.state == MqttConnectionState.connected) {
      print('MQTT client connected');
      _client.subscribe('vision/orientation', MqttQos.atMostOnce);
      _client.subscribe('vision/leds', MqttQos.atMostOnce);
      _client.subscribe('vision/circles', MqttQos.atMostOnce);
    } else {
      print(
          'MQTT client connection failed - disconnecting, status is ${_client.connectionStatus}');
      _client.disconnect();
    }

    _client.updates!.listen((List<MqttReceivedMessage<MqttMessage>> c) {
      final MqttPublishMessage recMess = c[0].payload as MqttPublishMessage;
      final String message =
          MqttPublishPayload.bytesToStringAsString(recMess.payload.message);
      print("TOPIC");
      print(c[0].topic);
      if (c[0].topic == 'vision/orientation') {
        setState(() {
          _mqttRotationFactor = double.parse(message);
          print(2 * math.pi * (_mqttRotationFactor / 56));
          print("ROTATION");
          print(_mqttRotationFactor);
        });
      } else if (c[0].topic == 'vision/leds') {
        /*setState(() {
          _mqttRadiusFactor = double.parse(message);
        });*/
      } else if (c[0].topic == 'vision/circles') {
        /*setState(() {
        _mqttRadiusFactor = double.parse(message);
      });*/
      }
    });
  }

  void _onSubscribed(String topic) {
    print('Subscription confirmed for topic $topic');
  }

  Future<void> _onDisconnected() async {
    print('MQTT client disconnected');
    try {
      await _client.connect();
    } catch (e) {
      print('Exception: $e');
      _client.disconnect();
    }
  }

  void _onConnected() {
    print('MQTT client connected');
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return CustomPaint(
          size: Size(300, 300),
          painter: ShapePainter(
            rotationAngle: 2 *
                math.pi *
                (_mqttRotationFactor /
                    56), //_rotationAnimation.value * _mqttRotationFactor,
            radiusFactor: _radiusAnimation.value * _mqttRadiusFactor,
            dotColors: _dotColors,
          ),
        );
      },
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    _client.disconnect();
    super.dispose();
  }
}

class ShapePainter extends CustomPainter {
  final double rotationAngle;
  final double radiusFactor;
  final List<Color> dotColors;

  ShapePainter(
      {required this.rotationAngle,
      required this.radiusFactor,
      required this.dotColors});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final maxRadius = size.width * 0.4;
    final currentRadius = maxRadius * radiusFactor;

    // Draw the orange arc
    final orangePaint = Paint()
      ..color = Colors.orange
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10;
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: currentRadius),
      -math.pi / 2,
      math.pi,
      false,
      orangePaint,
    );

    // Draw the red circle
    final redPaint = Paint()..color = Colors.red;
    canvas.drawCircle(center, 5, redPaint);

// Draw the rotating semicircle
    final semicircleRadius = maxRadius * 0.8;
    canvas.save();
    canvas.translate(center.dx, center.dy);
    canvas.rotate(rotationAngle);

    final semicirclePaint = Paint()
      ..color = Colors.purple.withOpacity(0.5)
      ..style = PaintingStyle.fill;

    canvas.drawArc(
      Rect.fromCircle(center: Offset(0, 0), radius: semicircleRadius),
      -math.pi / 2,
      math.pi,
      true,
      semicirclePaint,
    );
    canvas.restore();

    // Draw the 28 dots in a semicircle
    final dotRadius = size.width * 0.015;
    final dotCenterRadius = size.width * 0.6;
    for (int i = 0; i < 28; i++) {
      final angle = math.pi * i / 27; // Distribute over 180 degrees
      final dotCenter = Offset(
        center.dx + dotCenterRadius * math.cos(angle),
        center.dy + dotCenterRadius * math.sin(angle),
      );
      final dotPaint = Paint()..color = dotColors[i];
      canvas.drawCircle(dotCenter, dotRadius, dotPaint);
    }
  }

  @override
  bool shouldRepaint(covariant ShapePainter oldDelegate) {
    return oldDelegate.rotationAngle != rotationAngle ||
        oldDelegate.radiusFactor != radiusFactor ||
        oldDelegate.dotColors != dotColors;
  }
}
