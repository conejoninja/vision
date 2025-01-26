import 'dart:math';
import 'package:flutter/material.dart';

// Model for player position and rotation
class Player {
  double x;
  double y;
  double rotation; // in radians, 0 points right, increases clockwise

  Player({
    required this.x,
    required this.y,
    required this.rotation,
  });
}

// Custom painter for the maze
class MazePainter extends CustomPainter {
  final List<List<bool>> maze; // true represents a wall
  final Player player;
  static const double tileSize = 15.0;
  static const double playerSize = 6.0; // diameter of player dot
  static const double viewRadius = 15.0;
  static const double viewAngle = pi; // 180 degrees in radians

  MazePainter({
    required this.maze,
    required this.player,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final Paint gridPaint = Paint()
      ..color = Colors.grey
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.0;

    final Paint wallPaint = Paint()
      ..color = Colors.black
      ..style = PaintingStyle.fill;

    final Paint playerPaint = Paint()
      ..color = Colors.red
      ..style = PaintingStyle.fill;

    final Paint viewPaint = Paint()
      ..color = Colors.red.withOpacity(0.8)
      ..style = PaintingStyle.fill;

    // Draw grid and walls
    for (int y = 0; y < maze.length; y++) {
      for (int x = 0; x < maze[y].length; x++) {
        final rect = Rect.fromLTWH(
          x * tileSize,
          y * tileSize,
          tileSize,
          tileSize,
        );

        // Draw grid cell
        canvas.drawRect(rect, gridPaint);

        // Fill walls
        if (maze[y][x]) {
          canvas.drawRect(rect, wallPaint);
        }
      }
    }

    // Draw field of view
    final viewPath = Path();
    final centerX = player.x / 20; // * tileSize + tileSize / 2;
    final centerY = player.y / 20; // * tileSize + tileSize / 2;

    viewPath.moveTo(centerX, centerY);

    // Draw arc for field of view
    viewPath.arcTo(
      Rect.fromCircle(
        center: Offset(centerX, centerY),
        radius: viewRadius,
      ),
      player.rotation - viewAngle / 2, // start angle
      viewAngle, // sweep angle
      true,
    );

    viewPath.close();
    canvas.drawPath(viewPath, viewPaint);

    // Draw player
    canvas.drawCircle(
      Offset(centerX, centerY),
      playerSize / 2,
      playerPaint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}

// Example usage widget
class MazeView extends StatelessWidget {
  final List<List<bool>> maze;
  final Player player;

  const MazeView({
    super.key,
    required this.maze,
    required this.player,
  });

  @override
  Widget build(BuildContext context) {
    return CustomPaint(
      painter: MazePainter(
        maze: maze,
        player: player,
      ),
      // Size for 32x32 maze
      size: const Size(32 * MazePainter.tileSize, 32 * MazePainter.tileSize),
    );
  }
}
