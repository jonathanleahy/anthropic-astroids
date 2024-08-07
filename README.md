# Asteroids Game

This is a modern take on the classic Asteroids game, implemented in Go using the Ebiten 2D game library. The game was iteratively created through a series of prompts and responses with an AI assistant over the course of a couple of hours.

![Asteroids Game Demo](./astroids-video.gif)

## Features

- Classic Asteroids gameplay with a twist
- Smooth ship controls with realistic physics
- Particle effects for enhanced visual appeal
- Special expanding asteroids for added challenge
- Parallax star field background
- Score tracking and game over state

## How to Play

1. Use the arrow keys to control your ship:
    - Left/Right arrows to rotate
    - Up arrow for thrust
2. Press the Space bar to shoot bullets
3. Destroy asteroids to earn points
4. Avoid colliding with asteroids
5. Special golden asteroids will expand and contract before exploding into particles
6. The game ends when all asteroids are destroyed

## Installation

1. Ensure you have Go installed on your system
2. Install the Ebiten library:
   ```
   go get github.com/hajimehoshi/ebiten/v2
   ```
3. Clone this repository or copy the game code into a new Go file
4. Run the game:
   ```
   go run main.go
   ```

## Dependencies

- [Ebiten v2](https://github.com/hajimehoshi/ebiten)
- Go's standard library

## Development Process

This game was created through an iterative process using AI-assisted prompting over the course of a couple of hours. The initial concept and basic implementation were outlined, and then additional features and refinements were added based on ongoing dialogue and suggestions.

While this version doesn't implement every feature of the original Asteroids game, it captures the essence of the classic gameplay with some modern twists. The rapid development process demonstrates the potential of AI-assisted game creation.

## Future Improvements

- Add levels with increasing difficulty
- Implement power-ups and different weapon types
- Create a high score system
- Add sound effects and background music
- Enhance graphics with more detailed sprites
- Implement missing features from the original Asteroids game

## Credits

Developed with assistance from an AI language model, demonstrating the potential of AI-assisted game development.

## License

This project is open source and available under the MIT License.
