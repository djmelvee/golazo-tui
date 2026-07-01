package ascii

import (
	"math"
	"strings"
)

// ballShade maps surface brightness to ASCII (dark → bright).
const ballShade = " .'`^\",:;!iIlL|/\\|)(][}{*#"

// SpinBall3D renders one large soccer ball rotating in 3D for the splash screen.
func SpinBall3D(frame int) []string {
	const w, h = 44, 24
	zbuf := make([][]float64, h)
	out := make([][]rune, h)
	for y := 0; y < h; y++ {
		zbuf[y] = make([]float64, w)
		out[y] = make([]rune, w)
		for x := 0; x < w; x++ {
			zbuf[y][x] = math.Inf(-1)
			out[y][x] = ' '
		}
	}

	cx, cy := float64(w)/2, float64(h)/2
	ay := float64(frame) * math.Pi / 12
	ax := math.Pi/6 + math.Sin(float64(frame)*0.09)*0.15

	drawSoccerBall(zbuf, out, cx, cy, 8.5, ay, ax)

	lines := make([]string, h)
	for y := 0; y < h; y++ {
		lines[y] = strings.TrimRight(string(out[y]), " ")
	}
	return lines
}

func drawSoccerBall(zbuf [][]float64, out [][]rune, cx, cy, r, ay, ax float64) {
	lightX, lightY, lightZ := -0.35, 0.55, -0.75
	lLen := math.Hypot(lightX, math.Hypot(lightY, lightZ))
	lightX, lightY, lightZ = lightX/lLen, lightY/lLen, lightZ/lLen

	// Surface fill with pentagon/hexagon soccer pattern.
	for theta := 0.0; theta < 2*math.Pi; theta += 0.038 {
		for phi := 0.04; phi < math.Pi-0.04; phi += 0.075 {
			x := r * math.Sin(phi) * math.Cos(theta)
			y := r * math.Sin(phi) * math.Sin(theta)
			z := r * math.Cos(phi)

			rx, ry, rz := rotY(x, y, z, ay)
			rx, ry, rz = rotX(rx, ry, rz, ax)
			if rz < -r*0.15 {
				continue
			}

			sx, sy, ok := project(cx, cy, rx, ry, rz)
			if !ok || sy < 0 || sy >= len(out) || sx < 0 || sx >= len(out[0]) {
				continue
			}
			if rz <= zbuf[sy][sx] {
				continue
			}
			zbuf[sy][sx] = rz

			nx := math.Sin(phi) * math.Cos(theta)
			ny := math.Sin(phi) * math.Sin(theta)
			nz := math.Cos(phi)
			nx, ny, nz = rotY(nx, ny, nz, ay)
			nx, ny, nz = rotX(nx, ny, nz, ax)
			nLen := math.Hypot(nx, math.Hypot(ny, nz))
			dot := (nx/nLen)*lightX + (ny/nLen)*lightY + (nz/nLen)*lightZ
			if dot < 0 {
				dot = 0
			}

			pent := isPentagonPatch(theta, phi, ay)
			idx := int(dot * float64(len(ballShade)-1))
			if pent {
				idx = clampInt(idx-4, 0, len(ballShade)-1)
				if dot > 0.45 {
					out[sy][sx] = '#'
					continue
				}
			}
			out[sy][sx] = rune(ballShade[idx])
		}
	}

	// Seam lines for recognizable ball structure.
	for theta := 0.0; theta < 2*math.Pi; theta += 0.09 {
		for phi := 0.1; phi < math.Pi-0.1; phi += 0.09 {
			if !isSeam(theta, phi, ay) {
				continue
			}
			x := r * math.Sin(phi) * math.Cos(theta)
			y := r * math.Sin(phi) * math.Sin(theta)
			z := r * math.Cos(phi)
			rx, ry, rz := rotY(x, y, z, ay)
			rx, ry, rz = rotX(rx, ry, rz, ax)
			if rz < 0 {
				continue
			}
			sx, sy, ok := project(cx, cy, rx, ry, rz)
			if !ok || sy < 0 || sy >= len(out) || sx < 0 || sx >= len(out[0]) {
				continue
			}
			if out[sy][sx] == ' ' {
				out[sy][sx] = '-'
			}
		}
	}
}

func isPentagonPatch(theta, phi, ay float64) bool {
	u := math.Sin(5*(theta+ay*0.25)+0.4) * math.Cos(3*phi-0.2)
	v := math.Cos(4*theta) * math.Sin(2*phi+ay*0.15)
	return u+v > 0.72
}

func isSeam(theta, phi, ay float64) bool {
	a := math.Abs(math.Sin(5*theta+ay*0.2)) > 0.96
	b := math.Abs(math.Cos(6*phi)) > 0.97
	return a || b
}

func project(cx, cy, rx, ry, rz float64) (sx, sy int, ok bool) {
	depth := 5.0 + rz*0.35
	sx = int(cx + rx/depth*9)
	sy = int(cy + ry/depth*4.6)
	ok = true
	return sx, sy, ok
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func rotY(x, y, z, a float64) (float64, float64, float64) {
	c, s := math.Cos(a), math.Sin(a)
	return x*c + z*s, y, -x*s + z*c
}

func rotX(x, y, z, a float64) (float64, float64, float64) {
	c, s := math.Cos(a), math.Sin(a)
	return x, y*c - z*s, y*s + z*c
}