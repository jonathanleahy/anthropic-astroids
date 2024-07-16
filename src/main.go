package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"
)

const (
	screenWidth                  = 640
	screenHeight                 = 480
	shipSize                     = 20
	asteroidSize                 = 30
	maxVelocity                  = 5
	acceleration                 = 0.1
	friction                     = 0.99
	bulletSpeed                  = 10
	bulletLifetime               = 60
	numStars                     = 200
	numStarLayers                = 5
	numAsteroidVertices          = 10
	specialAsteroidProb          = 0.1
	specialAsteroidExpansionTime = 180 // 3 seconds at 60 FPS
)

type GameState int

const (
	StatePlaying GameState = iota
	StateGameOver
)

type Game struct {
	ship      Ship
	asteroids []Asteroid
	bullets   []Bullet
	stars     []Star
	particles []Particle
	score     int
	state     GameState
	font      font.Face
	worldX    float64
	worldY    float64
	tick      int
}

type Ship struct {
	angle, dx, dy float64
	thrusting     bool
}

type AsteroidType int

const (
	NormalAsteroid AsteroidType = iota
	SpecialAsteroid
)

type Asteroid struct {
	x, y, dx, dy, size, rotation, rotationSpeed float64
	points                                      []float64
	asteroidType                                AsteroidType
	expansionTime                               int
}

type Bullet struct {
	x, y, dx, dy float64
	lifetime     int
}

type Star struct {
	x, y       float64
	brightness float64
	layer      int
}

type Particle struct {
	x, y, dx, dy float64
	color        color.RGBA
	lifetime     float64
}

var emptySubImage = ebiten.NewImage(3, 3).SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

func (g *Game) Update() error {
	g.tick++
	switch g.state {
	case StatePlaying:
		g.updatePlaying()
	case StateGameOver:
		g.updateGameOver()
	}
	return nil
}

func (g *Game) updatePlaying() {
	// Ship rotation
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.ship.angle -= 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.ship.angle += 0.1
	}

	// Ship thrust
	g.ship.thrusting = ebiten.IsKeyPressed(ebiten.KeyUp)
	if g.ship.thrusting {
		g.ship.dx += math.Cos(g.ship.angle) * acceleration
		g.ship.dy += math.Sin(g.ship.angle) * acceleration
	}

	// Apply friction and limit speed
	g.ship.dx *= friction
	g.ship.dy *= friction
	speed := math.Sqrt(g.ship.dx*g.ship.dx + g.ship.dy*g.ship.dy)
	if speed > maxVelocity {
		g.ship.dx = g.ship.dx / speed * maxVelocity
		g.ship.dy = g.ship.dy / speed * maxVelocity
	}

	// Update world position
	g.worldX += g.ship.dx
	g.worldY += g.ship.dy

	// Update stars with parallax effect
	for i := range g.stars {
		parallaxFactor := float64(g.stars[i].layer) / float64(numStarLayers)
		g.stars[i].x -= g.ship.dx * parallaxFactor
		g.stars[i].y -= g.ship.dy * parallaxFactor
		g.stars[i].x = math.Mod(g.stars[i].x+screenWidth, screenWidth)
		g.stars[i].y = math.Mod(g.stars[i].y+screenHeight, screenHeight)
	}

	// Firing bullets
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.tick%10 == 0 {
		g.bullets = append(g.bullets, Bullet{
			x:        screenWidth/2 + math.Cos(g.ship.angle)*shipSize,
			y:        screenHeight/2 + math.Sin(g.ship.angle)*shipSize,
			dx:       math.Cos(g.ship.angle) * bulletSpeed,
			dy:       math.Sin(g.ship.angle) * bulletSpeed,
			lifetime: bulletLifetime,
		})
	}

	// Update bullets
	for i := 0; i < len(g.bullets); i++ {
		g.bullets[i].x += g.bullets[i].dx - g.ship.dx
		g.bullets[i].y += g.bullets[i].dy - g.ship.dy
		g.bullets[i].lifetime--
		if g.bullets[i].lifetime <= 0 {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
			i--
		}
	}

	// Update asteroids and check for collisions
	for i := 0; i < len(g.asteroids); i++ {
		a := &g.asteroids[i]
		a.x += a.dx - g.ship.dx
		a.y += a.dy - g.ship.dy
		a.rotation += a.rotationSpeed

		// Wrap asteroids around the screen
		a.x = math.Mod(a.x+screenWidth, screenWidth)
		a.y = math.Mod(a.y+screenHeight, screenHeight)
		if a.x < 0 {
			a.x += screenWidth
		}
		if a.y < 0 {
			a.y += screenHeight
		}

		if a.asteroidType == SpecialAsteroid {
			a.expansionTime--
			a.size = asteroidSize * (1 + math.Sin(float64(a.expansionTime)/specialAsteroidExpansionTime*math.Pi)*0.5)

			if a.expansionTime <= 0 {
				// Explode into particles
				for j := 0; j < 100; j++ {
					angle := rand.Float64() * 2 * math.Pi
					speed := rand.Float64() * 5
					particle := Particle{
						x:  a.x,
						y:  a.y,
						dx: math.Cos(angle) * speed,
						dy: math.Sin(angle) * speed,
						color: color.RGBA{
							R: 255,
							G: 255,
							B: 0,
							A: 255,
						},
						lifetime: 3 * 60, // 3 seconds at 60 FPS
					}
					g.particles = append(g.particles, particle)
				}
				// Remove the special asteroid
				g.asteroids = append(g.asteroids[:i], g.asteroids[i+1:]...)
				i--
				continue
			}
		}

		for j := 0; j < len(g.bullets); j++ {
			if distance(a.x, a.y, g.bullets[j].x, g.bullets[j].y) < a.size/2 {
				// Split asteroid
				if a.size > 25 {
					g.asteroids = append(g.asteroids, createAsteroid(a.x, a.y, a.size/2))
					*a = createAsteroid(a.x, a.y, a.size/2)
				} else {
					// Remove small asteroid
					g.asteroids = append(g.asteroids[:i], g.asteroids[i+1:]...)
					i--
				}
				// Remove bullet and increase score
				g.bullets = append(g.bullets[:j], g.bullets[j+1:]...)
				g.score += 10
				break
			}
		}
	}

	// Update particles
	for i := 0; i < len(g.particles); i++ {
		p := &g.particles[i]
		p.x += p.dx
		p.y += p.dy
		p.lifetime--

		// Cool down the particle color
		p.color.R = uint8(float64(p.color.R) * (p.lifetime / (3 * 60)))
		p.color.G = uint8(float64(p.color.G) * (p.lifetime / (3 * 60)))

		if p.lifetime <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
			i--
		}
	}

	// Check if all asteroids are destroyed
	if len(g.asteroids) == 0 {
		g.state = StateGameOver
	}
}

func (g *Game) updateGameOver() {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.restartGame()
	}
}

func (g *Game) restartGame() {
	g.ship = Ship{}
	g.asteroids = make([]Asteroid, 5)
	g.bullets = []Bullet{}
	g.particles = []Particle{}
	g.score = 0
	g.state = StatePlaying
	g.worldX = 0
	g.worldY = 0

	for i := range g.asteroids {
		g.asteroids[i] = createAsteroid(float64(rand.Intn(screenWidth)), float64(rand.Intn(screenHeight)), asteroidSize)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw stars with parallax effect
	for _, s := range g.stars {
		starColor := color.RGBA{
			R: uint8(s.brightness * float64(s.layer) / float64(numStarLayers) * 255),
			G: uint8(s.brightness * float64(s.layer) / float64(numStarLayers) * 255),
			B: uint8(s.brightness * float64(s.layer) / float64(numStarLayers) * 255),
			A: 255,
		}
		ebitenutil.DrawRect(screen, s.x, s.y, 1, 1, starColor)
	}

	// Draw asteroids
	for _, a := range g.asteroids {
		drawAsteroid(screen, a)
	}

	// Draw ship (centered)
	shipPoints := []float64{
		screenWidth/2 + math.Cos(g.ship.angle)*shipSize, screenHeight/2 + math.Sin(g.ship.angle)*shipSize,
		screenWidth/2 + math.Cos(g.ship.angle+2.6)*shipSize*0.7, screenHeight/2 + math.Sin(g.ship.angle+2.6)*shipSize*0.7,
		screenWidth/2 + math.Cos(g.ship.angle-2.6)*shipSize*0.7, screenHeight/2 + math.Sin(g.ship.angle-2.6)*shipSize*0.7,
	}
	ebitenutil.DrawLine(screen, shipPoints[0], shipPoints[1], shipPoints[2], shipPoints[3], color.White)
	ebitenutil.DrawLine(screen, shipPoints[2], shipPoints[3], shipPoints[4], shipPoints[5], color.White)
	ebitenutil.DrawLine(screen, shipPoints[4], shipPoints[5], shipPoints[0], shipPoints[1], color.White)

	// Draw thrust
	if g.ship.thrusting {
		thrustPoints := []float64{
			screenWidth/2 - math.Cos(g.ship.angle)*shipSize*0.5, screenHeight/2 - math.Sin(g.ship.angle)*shipSize*0.5,
			screenWidth/2 - math.Cos(g.ship.angle)*shipSize*0.8, screenHeight/2 - math.Sin(g.ship.angle)*shipSize*0.8,
		}
		ebitenutil.DrawLine(screen, thrustPoints[0], thrustPoints[1], thrustPoints[2], thrustPoints[3], color.RGBA{255, 165, 0, 255})
	}

	// Draw bullets
	for _, b := range g.bullets {
		ebitenutil.DrawCircle(screen, b.x, b.y, 2, color.White)
	}

	// Draw particles
	for _, p := range g.particles {
		ebitenutil.DrawRect(screen, p.x, p.y, 2, 2, p.color)
	}

	// Draw score
	scoreText := fmt.Sprintf("Score: %d", g.score)
	text.Draw(screen, scoreText, g.font, 10, 20, color.White)

	// Draw speed
	speed := math.Sqrt(g.ship.dx*g.ship.dx + g.ship.dy*g.ship.dy)
	speedText := fmt.Sprintf("Speed: %.2f", speed)
	text.Draw(screen, speedText, g.font, 10, 50, color.White)

	if g.state == StateGameOver {
		gameOverText := "Game Over! Press SPACE to restart"
		textBounds := text.BoundString(g.font, gameOverText)
		text.Draw(screen, gameOverText, g.font, screenWidth/2-textBounds.Dx()/2, screenHeight/2, color.White)
	}
}

func createAsteroid(x, y, size float64) Asteroid {
	asteroidType := NormalAsteroid
	if rand.Float64() < specialAsteroidProb {
		asteroidType = SpecialAsteroid
	}

	asteroid := Asteroid{
		x:             x,
		y:             y,
		dx:            (rand.Float64()*2 - 1) * 2, // Increased speed
		dy:            (rand.Float64()*2 - 1) * 2, // Increased speed
		size:          size,
		rotation:      0,
		rotationSpeed: (rand.Float64() - 0.5) * 0.1, // Increased rotation speed
		points:        make([]float64, numAsteroidVertices*2),
		asteroidType:  asteroidType,
		expansionTime: specialAsteroidExpansionTime,
	}

	for i := 0; i < numAsteroidVertices; i++ {
		angle := float64(i) * (2 * math.Pi / numAsteroidVertices)
		r := size/2 + rand.Float64()*size/4 - size/8
		asteroid.points[i*2] = math.Cos(angle) * r
		asteroid.points[i*2+1] = math.Sin(angle) * r
	}

	return asteroid
}

func drawAsteroid(screen *ebiten.Image, a Asteroid) {
	asteroidColor := color.RGBA{128, 128, 128, 255}
	if a.asteroidType == SpecialAsteroid {
		pulse := float64(a.expansionTime) / specialAsteroidExpansionTime
		asteroidColor = color.RGBA{
			R: uint8(128 + 127*math.Sin(pulse*math.Pi)),
			G: uint8(128 - 64*math.Sin(pulse*math.Pi)),
			B: uint8(128 - 64*math.Sin(pulse*math.Pi)),
			A: 255,
		}
	}

	var path vector.Path
	path.MoveTo(float32(a.x+a.points[0]*math.Cos(a.rotation)-a.points[1]*math.Sin(a.rotation)),
		float32(a.y+a.points[0]*math.Sin(a.rotation)+a.points[1]*math.Cos(a.rotation)))

	for i := 1; i < numAsteroidVertices; i++ {
		x := float32(a.x + a.points[i*2]*math.Cos(a.rotation) - a.points[i*2+1]*math.Sin(a.rotation))
		y := float32(a.y + a.points[i*2]*math.Sin(a.rotation) + a.points[i*2+1]*math.Cos(a.rotation))
		path.LineTo(x, y)
	}
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)

	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(asteroidColor.R) / 255
		vs[i].ColorG = float32(asteroidColor.G) / 255
		vs[i].ColorB = float32(asteroidColor.B) / 255
		vs[i].ColorA = float32(asteroidColor.A) / 255
	}

	op := &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.EvenOdd,
	}

	screen.DrawTriangles(vs, is, emptySubImage, op)

	// Draw the center point and outline
	//ebitenutil.DrawCircle(screen, a.x, a.y, 2, color.RGBA{255, 0, 0, 255})
	for i := 0; i < numAsteroidVertices; i++ {
		j := (i + 1) % numAsteroidVertices
		x1 := a.x + a.points[i*2]*math.Cos(a.rotation) - a.points[i*2+1]*math.Sin(a.rotation)
		y1 := a.y + a.points[i*2]*math.Sin(a.rotation) + a.points[i*2+1]*math.Cos(a.rotation)
		x2 := a.x + a.points[j*2]*math.Cos(a.rotation) - a.points[j*2+1]*math.Sin(a.rotation)
		y2 := a.y + a.points[j*2]*math.Sin(a.rotation) + a.points[j*2+1]*math.Cos(a.rotation)
		ebitenutil.DrawLine(screen, x1, y1, x2, y2, asteroidColor)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Load font
	fontData, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	font, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		asteroids: make([]Asteroid, 15),
		stars:     make([]Star, numStars),
		state:     StatePlaying,
		font:      font,
	}

	for i := range game.asteroids {
		game.asteroids[i] = createAsteroid(float64(rand.Intn(screenWidth)), float64(rand.Intn(screenHeight)), asteroidSize)
	}

	for i := range game.stars {
		game.stars[i] = Star{
			x:          float64(rand.Intn(screenWidth)),
			y:          float64(rand.Intn(screenHeight)),
			brightness: rand.Float64()*0.5 + 0.5,
			layer:      rand.Intn(numStarLayers) + 1,
		}
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Asteroids")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
