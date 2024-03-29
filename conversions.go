// Copyright © 2019 Xavier Basty <xavier@hexbee.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gocolor

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Hexbee-net/gocolor/named"
)

////////////////////////////////////////

// RGBtoHSL converts a color from base RGB coordinates to HSL.
func RGBtoHSL(r, g, b float64) (h, s, l float64, err error) {
	if err := checkRGB(r, g, b); err != nil {
		return 0, 0, 0, err
	}

	minVal := min(r, g, b)
	maxVal := max(r, g, b)

	l = (maxVal + minVal) / 2
	if minVal == maxVal {
		return 0, 0, l, nil // Achromatic (gray)
	}

	d := maxVal - minVal // delta RGB value

	if l < 0.5 {
		s = d / (maxVal + minVal)
	} else {
		s = d / (2 - maxVal - minVal)
	}

	dr := maxVal - r/d
	dg := maxVal - g/d
	db := maxVal - b/d

	if r == maxVal {
		h = db - dg
	} else if g == maxVal {
		h = 2 + dr - db
	} else {
		h = 4 + dg - dr
	}

	if h > 0 {
		h = math.Mod(h*60, 360)
	} else {
		h = math.Mod(360-math.Abs(h*60), 360)
	}

	return h, s, l, nil
}

// RGBtoHSV converts a color from base RGB coordinates to HSV.
func RGBtoHSV(r, g, b float64) (h, s, v float64, err error) {
	if err := checkRGB(r, g, b); err != nil {
		return 0, 0, 0, err
	}

	v = max(r, g, b)
	d := v - min(r, g, b)
	if d == 0 {
		return 0, 0, v, nil
	}

	s = d / v

	dr := (v - r) / d
	dg := (v - g) / d
	db := (v - b) / d

	if v == r {
		// between yellow & magenta
		h = db - dg
	} else if v == g {
		// between cyan & yellow
		h = 2 + dr - db
	} else { // v == b
		// between magenta & cyan
		h = 4 + dg - dr
	}

	if h > 0 {
		h = math.Mod(h*60, 360)
	} else {
		h = math.Mod(360-math.Abs(h*60), 360)
	}

	return h, s, v, nil
}

// RGBtoYIQ converts a color from base RGB coordinates to YIQ (Luma In-phase Quadrature).
func RGBtoYIQ(r, g, b float64) (y, i, q float64, err error) {
	if err := checkRGB(r, g, b); err != nil {
		return 0, 0, 0, err
	}

	yiq := conversionRgbYiq.vdot(vector{r, g, b})
	return yiq.v0, yiq.v1, yiq.v2, nil
}

// RGBtoSDYUV converts a color from base RGB coordinates to SDTV YUV (BT.601).
func RGBtoSDYUV(r, g, b float64) (y, u, v float64, err error) {
	if err := checkRGB(r, g, b); err != nil {
		return 0, 0, 0, err
	}

	yuv := conversionRgbSDYuv.vdot(vector{r, g, b})
	return yuv.v0, yuv.v1, yuv.v2, nil
}

// RGBtoHDYUV converts a color from base RGB coordinates to HDTV YUV (BT.709).
func RGBtoHDYUV(r, g, b float64) (y, u, v float64, err error) {
	if err := checkRGB(r, g, b); err != nil {
		return 0, 0, 0, err
	}

	yuv := conversionRgbHDYuv.vdot(vector{r, g, b})
	return yuv.v0, yuv.v1, yuv.v2, nil
}

// RGBtoCMY converts a color from base RGB coordinates to CMY.
func RGBtoCMY(r, g, b float64) (float64, float64, float64, error) {
	if err := checkRGB(r, g, b); err != nil {
		return 0, 0, 0, err
	}

	return 1 - r, 1 - g, 1 - b, nil
}

// RGBtoHEX converts a color from base RGB coordinates to #RRGGBB.
func RGBtoHEX(r, g, b float64) (string, error) {
	if err := checkRGB(r, g, b); err != nil {
		return "", err
	}

	ri := int(math.Min(math.Round(r*255), 255))
	gi := int(math.Min(math.Round(g*255), 255))
	bi := int(math.Min(math.Round(b*255), 255))
	return fmt.Sprintf("#%02X%02X%02X", ri, gi, bi), nil
}

// RGBtoXYZ converts a color from RGB coordinates to XYZ.
// The illuminant for the XYZ color is D65 and the observer's angle 2°.
func RGBtoXYZ(r, g, b float64, space string) (x, y, z float64, err error) {
	if err := checkRGB(r, g, b); err != nil {
		return 0, 0, 0, err
	}

	switch space {
	case SRGB:
		linearize := func(v float64) float64 {
			if v <= 0.04045 {
				return v / 12.92
			}
			return math.Pow((v+0.055)/1.055, 2.4)
		}
		r = linearize(r)
		g = linearize(g)
		b = linearize(b)

	case BT2020:
		linearize := func(v float64) float64 {
			if v <= 0.08124794403514049 {
				return v / 4.5
			}
			return math.Pow((v+0.099)/1.099, 1/0.45)
		}
		r = linearize(r)
		g = linearize(g)
		b = linearize(b)

	case BT202012b:
		linearize := func(v float64) float64 {
			if v <= 0.081697877417347 {
				return v / 4.5
			}
			return math.Pow((v+0.0993)/1.0993, 1/0.45)
		}
		r = linearize(r)
		g = linearize(g)
		b = linearize(b)

	default:
		gamma, ok := RGBGamma[space]
		if !ok {
			return 0, 0, 0, fmt.Errorf("could not find gamma for RGB color space: %v", space)
		}
		r = math.Pow(r, gamma)
		g = math.Pow(g, gamma)
		b = math.Pow(b, gamma)
	}

	m, ok := conversionRgbXyz[space]
	if !ok {
		return 0, 0, 0, fmt.Errorf("could not find conversion matrix for RGB color space: %v", space)
	}

	v := m.vdot(vector{r, g, b})
	return math.Max(v.v0, 0), math.Max(v.v1, 0), math.Max(v.v2, 0), nil
}

////////////////////////////////////////

// HSLtoRGB converts a color from HSL coordinates to RGB.
//
// The hue is in degrees and expected to be in the [0.0-360.0] range.
// The saturation and lightness are percentages in the [0.0-1.0] range.
func HSLtoRGB(h, s, l float64) (r, g, b float64, err error) {
	if h < 0 || h > 360 {
		return 0, 0, 0, fmt.Errorf("hue (H) is out of the [0, 360] range (%v)", h)
	}
	if s < 0 || s > 1 {
		return 0, 0, 0, fmt.Errorf("saturation (S) is out of the [0, 1] range (%v)", s)
	}
	if l < 0 || l > 1 {
		return 0, 0, 0, fmt.Errorf("lightness (L) is out of the [0, 1] range (%v)", l)
	}

	if s == 0 {
		return l, l, l, nil // achromatic (gray)
	}

	var q, p float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - (l * s)
	}
	p = 2*l - q

	hueToRgb := func(t float64) float64 {
		if t < 0 {
			t = t + 1
		}
		if t > 1 {
			t = t - 1
		}

		switch {
		case t < 1.0/6.0:
			return p + (q-p)*6*t
		case t < 1.0/2.0:
			return q
		case t < 2.0/3.0:
			return p + (q-p)*(2.0/3.0-t)*6.0
		default:
			return p
		}
	}

	k := h / 360.0
	r = hueToRgb(k + 1.0/3.0)
	g = hueToRgb(k)
	b = hueToRgb(k - 1.0/3.0)

	return r, g, b, nil
}

// HSVtoRGB converts a color from HSV coordinates to RGB.
//
// The hue is in degrees and expected to be in the [0.0-360.0] range.
// The saturation and value are percentages in the [0.0-1.0] range.
func HSVtoRGB(h, s, v float64) (r, g, b float64, err error) {
	if h < 0 || h > 360 {
		return 0, 0, 0, fmt.Errorf("hue (H) is out of the [0, 360] range (%v)", h)
	}
	if s < 0 || s > 1 {
		return 0, 0, 0, fmt.Errorf("saturation (S) is out of the [0, 1] range (%v)", s)
	}
	if v < 0 || v > 1 {
		return 0, 0, 0, fmt.Errorf("value (V) is out of the [0, 1] range (%v)", v)
	}

	if s == 0 {
		return v, v, v, nil // achromatic (gray)
	}

	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	switch int(math.Mod(h/60, 6)) {
	case 0: // 0º <= h < 60º
		r, g, b = c, x, 0
	case 1: // 60º <= h < 120º
		r, g, b = x, c, 0
	case 2: // 120º <= h < 180º
		r, g, b = 0, c, x
	case 3: // 180 <= h < 240º
		r, g, b = 0, x, c
	case 4: // 240º <= h < 300º
		r, g, b = x, 0, c
	case 5: // 300º <= h < 360
		r, g, b = c, 0, x
	}

	return r + m, g + m, b + m, nil
}

// YIQtoRGB converts a color from YIQ coordinates to RGB.
func YIQtoRGB(y, i, q float64) (r, g, b float64, err error) {
	if y < 0 || y > 1 {
		return 0, 0, 0, fmt.Errorf("luma (Y) is out of the [0, 1] range (%v)", y)
	}
	if i < -0.523 || i > 0.523 {
		return 0, 0, 0, fmt.Errorf("in-phase (I) is out of the [-0.523, 0.523] range (%v)", i)
	}
	if q < -0.596 || q > 0.596 {
		return 0, 0, 0, fmt.Errorf("quadrature (Q) is out of the [-0.596, 0.596] range (%v)", q)
	}

	rgb := conversionYiqRgb.vdot(vector{y, i, q})
	return rgb.v0, rgb.v1, rgb.v2, nil
}

// SDYUVtoRGB converts a color from YUV coordinates to RGB.
func SDYUVtoRGB(y, u, v float64) (r, g, b float64, err error) {
	if err = checkYUV(y, u, v); err != nil {
		return 0, 0, 0, err
	}

	rgb := conversionSDYuvRgb.vdot(vector{y, u, v})
	return rgb.v0, rgb.v1, rgb.v2, nil
}

// HDYUVtoRGB converts a color from YUV coordinates to RGB.
func HDYUVtoRGB(y, u, v float64) (r, g, b float64, err error) {
	if err = checkYUV(y, u, v); err != nil {
		return 0, 0, 0, err
	}

	rgb := conversionHDYuvRgb.vdot(vector{y, u, v})
	return rgb.v0, rgb.v1, rgb.v2, nil
}

// CMYtoRGB converts a color from CMY coordinates to RGB.
func CMYtoRGB(c, m, y float64) (float64, float64, float64, error) {
	if c < 0 || c > 1 {
		return 0, 0, 0, fmt.Errorf("cyan is out of the [0, 1] range (%v)", c)
	}
	if m < 0 || m > 1 {
		return 0, 0, 0, fmt.Errorf("magenta is out of the [0, 1] range (%v)", m)
	}
	if y < 0 || y > 1 {
		return 0, 0, 0, fmt.Errorf("yellow is out of the [0, 1] range (%v)", y)
	}

	return 1 - c, 1 - m, 1 - y, nil
}

// HEXtoRGB converts a color from HTML #RRGGBB to RGB coordinates.
func HEXtoRGB(html string) (r, g, b float64, err error) {
	html = strings.TrimSpace(html)
	if len(html) == 0 {
		return 0, 0, 0, errors.New("input is empty")
	}
	if html[0] == '#' {
		html = html[1:]
	} else {
		if val, ok := named.NamedColors[strings.ToLower(html)]; ok {
			html = val[1:]
		}
	}

	switch len(html) {
	// Long html code
	case 6:
		ri, err := strconv.ParseUint(html[0:2], 16, 64)
		if err != nil {
			return 0, 0, 0, err
		}
		r = float64(ri) / 255

		gi, err := strconv.ParseUint(html[2:4], 16, 64)
		if err != nil {
			return 0, 0, 0, err
		}
		g = float64(gi) / 255

		bi, err := strconv.ParseUint(html[4:6], 16, 64)
		if err != nil {
			return 0, 0, 0, err
		}
		b = float64(bi) / 255

	// Short html code
	case 3:
		ri, err := strconv.ParseUint(html[0:1], 16, 64)
		if err != nil {
			return 0, 0, 0, err
		}
		r = float64(ri*16+ri) / 255

		gi, err := strconv.ParseUint(html[1:2], 16, 64)
		if err != nil {
			return 0, 0, 0, err
		}
		g = float64(gi*16+gi) / 255

		bi, err := strconv.ParseUint(html[2:3], 16, 64)
		if err != nil {
			return 0, 0, 0, err
		}
		b = float64(bi*16+bi) / 255

	default:
		return 0, 0, 0, fmt.Errorf("input '%s' is not in #RRGGBB format", html)
	}

	return r, g, b, nil
}

// XYZtoRGB converts a color from XYZ coordinates to RGB.
// The illuminant for the XYZ color is assumed to be D65 and the observer's angle 2°.
func XYZtoRGB(x, y, z float64, space string) (r, g, b float64, err error) {
	if err := checkXYZ(x, y, z); err != nil {
		return 0, 0, 0, err
	}

	m, ok := conversionXyzRgb[space]
	if !ok {
		return 0, 0, 0, fmt.Errorf("unrecognized RGB color space: %v", space)
	}

	v := m.vdot(vector{x, y, z})
	r, g, b = v.v0, v.v1, v.v2

	switch space {
	case SRGB:
		delinearize := func(v float64) float64 {
			if v <= 0.0031308 {
				return v * 12.92
			}
			return 1.055*math.Pow(v, 1/2.4) - 0.055
		}
		r = delinearize(r)
		g = delinearize(g)
		b = delinearize(b)

	case BT2020:
		delinearize := func(v float64) float64 {
			if v < 0.018 {
				return v * 4.5
			}
			return 1.099*math.Pow(v, 0.45) - 0.099
		}
		r = delinearize(r)
		g = delinearize(g)
		b = delinearize(b)

	case BT202012b:
		delinearize := func(v float64) float64 {
			if v < 0.0181 {
				return v * 4.5
			}
			return 1.0993*math.Pow(v, 0.45) - 0.0993
		}
		r = delinearize(r)
		g = delinearize(g)
		b = delinearize(b)

	default:
		gamma, ok := RGBGamma[space]
		if !ok {
			return 0, 0, 0, fmt.Errorf("unrecognized RGB color space: %v", space)
		}
		r = math.Pow(r, 1/gamma)
		g = math.Pow(g, 1/gamma)
		b = math.Pow(b, 1/gamma)
	}

	return r, g, b, nil
}

////////////////////////////////////////

// CMYKtoCMY converts a color from CMYK coordinates to CMY.
func CMYKtoCMY(c, m, y, k float64) (float64, float64, float64, error) {
	if c < 0 || c > 1 {
		return 0, 0, 0, fmt.Errorf("cyan is out of the [0, 1] range (%v)", c)
	}
	if m < 0 || m > 1 {
		return 0, 0, 0, fmt.Errorf("magenta is out of the [0, 1] range (%v)", m)
	}
	if y < 0 || y > 1 {
		return 0, 0, 0, fmt.Errorf("yellow is out of the [0, 1] range (%v)", y)
	}
	if k < 0 || k > 1 {
		return 0, 0, 0, fmt.Errorf("black is out of the [0, 1] range (%v)", k)
	}

	mk := 1 - k
	return (c * mk) + k, (m * mk) + k, (y * mk) + k, nil
}

// CMYtoCMYK converts a color from CMY coordinates to CMYK.
func CMYtoCMYK(c, m, y float64) (float64, float64, float64, float64, error) {
	if c < 0 || c > 1 {
		return 0, 0, 0, 0, fmt.Errorf("cyan is out of the [0, 1] range (%v)", c)
	}
	if m < 0 || m > 1 {
		return 0, 0, 0, 0, fmt.Errorf("magenta is out of the [0, 1] range (%v)", m)
	}
	if y < 0 || y > 1 {
		return 0, 0, 0, 0, fmt.Errorf("yellow is out of the [0, 1] range (%v)", y)
	}

	k := min(c, m, y)
	if k == 1.0 {
		return 0.0, 0.0, 0.0, 1.0, nil
	}

	mk := 1 - k
	return (c - k) / mk, (m - k) / mk, (y - k) / mk, k, nil
}

////////////////////////////////////////

// XYZtoLAB converts a color from XYZ coordinates to Lab.
func XYZtoLAB(x, y, z float64, observer int, illuminant string) (l, a, b float64, err error) {
	if err := checkXYZ(x, y, z); err != nil {
		return 0, 0, 0, err
	}

	wp, err := getWhitePoint(observer, illuminant)
	if err != nil {
		return 0, 0, 0, err
	}

	x /= wp.v0
	y /= wp.v1
	z /= wp.v2

	if x > CieE {
		x = math.Pow(x, 1/3)
	} else {
		x = (7.787 * x) + (16.0 / 116.0)
	}

	if y > CieE {
		y = math.Pow(y, 1/3)
	} else {
		y = (7.787 * y) + (16.0 / 116.0)
	}

	if z > CieE {
		z = math.Pow(z, 1/3)
	} else {
		z = (7.787 * z) + (16.0 / 116.0)
	}

	l = (116.0 * y) - 16.0
	a = 500.0 * (x - y)
	b = 200.0 * (y - z)

	return l, a, b, nil
}

// XYZtoXYY converts a color from XYZ coordinates to xyY.
func XYZtoXYY(x, y, z float64) (float64, float64, float64, error) {
	if err := checkXYZ(x, y, z); err != nil {
		return 0, 0, 0, err
	}

	var xyyX, xyyY float64
	if s := x + y + z; s == 0 {
		xyyX = 0
		xyyY = 0
	} else {
		xyyX = x / s
		xyyY = y / s
	}
	return xyyX, xyyY, y, nil
}

// XYZtoLUV converts a color from XYZ coordinates to Luv.
func XYZtoLUV(x, y, z float64, observer int, illuminant string) (l, u, v float64, err error) {
	if err := checkXYZ(x, y, z); err != nil {
		return 0, 0, 0, err
	}

	wp, err := getWhitePoint(observer, illuminant)
	if err != nil {
		return 0, 0, 0, err
	}

	d := x + (15.0 * y) + (3.0 * z)
	if d == 0.0 {
		u = 0.0
		v = 0.0
	} else {
		u = (4.0 * x) / d
		v = (9.0 * y) / d
	}

	y = y / wp.v1
	if y > CieE {
		y = math.Pow(y, 1/3)
	} else {
		y = (7.787 * y) + (16.0 / 116.0)
	}

	refU := (4.0 * wp.v0) / (wp.v0 + (15.0 * wp.v1) + (3.0 * wp.v2))
	refV := (9.0 * wp.v1) / (wp.v0 + (15.0 * wp.v1) + (3.0 * wp.v2))

	l = (116.0 * y) - 16.0
	u = 13.0 * l * (u - refU)
	v = 13.0 * l * (v - refV)

	return l, u, v, nil
}

// XYYtoXYZ converts a color from xyZ coordinates to XYZ.
func XYYtoXYZ(x, y, Y float64) (float64, float64, float64, error) {
	if x < 0 || x > 1 {
		return 0, 0, 0, fmt.Errorf("x chrominance is out of the [0, 1] range (%v)", x)
	}
	if y < 0 || y > 1 {
		return 0, 0, 0, fmt.Errorf("y chrominance is out of the [0, 1] range (%v)", y)
	}
	if Y < 0 || Y > 1 {
		return 0, 0, 0, fmt.Errorf("luminance (Y) is out of the [0, 1] range (%v)", Y)
	}

	if y == 0 {
		return 0, 0, 0, nil
	}

	xyzX := (x * Y) / y
	xyzY := Y
	xyzZ := ((1.0 - x - y) * xyzY) / y

	return xyzX, xyzY, xyzZ, nil
}

// XYZtoIPT converts a color from XYZ coordinates to IPT.
func XYZtoIPT(x, y, z float64, observer int, illuminant string) (float64, float64, float64, error) {
	if err := checkXYZ(x, y, z); err != nil {
		return 0, 0, 0, err
	}

	if observer != Observer2 || illuminant != RefIlluminantD65 {
		return 0, 0, 0, errors.New("XYZ color for XYZ->IPT conversion needs to be D65 adapted")
	}
	prime := func(v float64) float64 {
		r := math.Pow(math.Abs(v), 0.43)
		if math.Signbit(v) {
			return -r
		}
		return r
	}

	lms := conversionXyzLms.vdot(vector{x, y, z})
	ipt := conversionLmsIpt.vdot(lms.mapfunc(prime))

	return ipt.v0, ipt.v1, ipt.v2, nil
}

////////////////////////////////////////

// LABtoXYZ converts a color from Lab coordinates to XYZ.
func LABtoXYZ(l, a, b float64, observer int, illuminant string) (x, y, z float64, err error) {
	if err := checkLAB(l, a, b); err != nil {
		return 0, 0, 0, err
	}

	wp, err := getWhitePoint(observer, illuminant)
	if err != nil {
		return 0, 0, 0, err
	}

	y = (l + 16) / 116
	x = a/500 + y
	z = y - b/200

	if px := math.Pow(x, 3); px > CieE {
		x = px
	} else {
		x = (x - 16/116) / 7.787
	}

	if py := math.Pow(y, 3); py > CieE {
		y = py
	} else {
		y = (y - 16/116) / 7.787
	}

	if pz := math.Pow(z, 3); pz > CieE {
		z = pz
	} else {
		z = (z - 16/116) / 7.787
	}

	x *= wp.v0
	y *= wp.v1
	z *= wp.v2

	return x, y, z, nil
}

// LUVtoXYZ converts a color from Luv coordinates to XYZ.
func LUVtoXYZ(l, u, v float64, observer int, illuminant string) (x, y, z float64, err error) {
	const cieKE = CieK * CieE

	if err := checkLUV(l, u, v); err != nil {
		return 0, 0, 0, err
	}

	if l <= 0.0 { // Without light, there is no color.
		return 0, 0, 0, nil
	}

	wp, err := getWhitePoint(observer, illuminant)
	if err != nil {
		return 0, 0, 0, err
	}

	vu := u/(13.0*l) + (4.0*wp.v0)/(wp.v0+15.0*wp.v1+3.0*wp.v2)
	vv := v/(13.0*l) + (9.0*wp.v1)/(wp.v0+15.0*wp.v1+3.0*wp.v2)

	// Y-coordinate calculations.
	if l > cieKE {
		y = math.Pow((l+16.0)/116.0, 3.0)
	} else {
		y = l / CieK
	}

	// X-coordinate calculation.
	x = y * 9.0 * vu / (4.0 * vv)

	// Z-coordinate calculation.
	z = y * (12.0 - 3.0*vu - 20.0*vv) / (4.0 * vv)

	return x, y, z, nil
}

// LABtoLCHAB converts a color from LAB coordinates to LCHab.
func LABtoLCHAB(l, a, b float64) (float64, float64, float64, error) {
	if err := checkLAB(l, a, b); err != nil {
		return 0, 0, 0, err
	}

	c := math.Sqrt(a*a + b*b)
	h := degrees(math.Atan2(b, a))

	return l, c, h, nil
}

// LCHABtoLAB converts a color from LCHab coordinates to LAB.
func LCHABtoLAB(l, c, h float64) (float64, float64, float64, error) {
	if h < 0 || h > 360 {
		return 0, 0, 0, fmt.Errorf("hue (h) is out of the [0, 360] range (%v)", h)
	}
	if c < -1 || c > 1 {
		return 0, 0, 0, fmt.Errorf("chroma (C) is out of the [0, 1] range (%v)", c)
	}
	if l < -1 || l > 1 {
		return 0, 0, 0, fmt.Errorf("lightness (L) is out of the [0, 1] range (%v)", l)
	}

	h = radians(h)
	a := math.Cos(h) * c
	b := math.Sin(h) * c

	return l, a, b, nil
}

// LUVtoLCHUV converts a color from Luv coordinates to LCHuv.
func LUVtoLCHUV(l, u, v float64) (float64, float64, float64, error) {
	if err := checkLUV(l, u, v); err != nil {
		return 0, 0, 0, err
	}

	c := math.Sqrt(u*u + v*v)
	h := degrees(math.Atan2(v, u))

	return l, c, h, nil
}

// LCHUVtoLUV converts a color from LCHuv coordinates to Luv.
func LCHUVtoLUV(l, c, h float64) (float64, float64, float64, error) {
	if h < 0 || h > 360 {
		return 0, 0, 0, fmt.Errorf("hue (h) is out of the [0, 360] range (%v)", h)
	}
	if c < -1 || c > 1 {
		return 0, 0, 0, fmt.Errorf("chroma (C) is out of the [0, 1] range (%v)", c)
	}
	if l < -1 || l > 1 {
		return 0, 0, 0, fmt.Errorf("lightness (L) is out of the [0, 1] range (%v)", l)
	}

	h = radians(h)
	u := math.Cos(h) * c
	v := math.Sin(h) * c

	return l, u, v, nil
}

////////////////////////////////////////

// IPTtoXYZ converts a color from IPT coordinates to XYZ.
func IPTtoXYZ(i, p, t float64) (float64, float64, float64, error) {
	if i < 0 || i > 1 {
		return 0, 0, 0, fmt.Errorf("i is out of the [0, 1] range (%v)", i)
	}
	if p < 0 || p > 1 {
		return 0, 0, 0, fmt.Errorf("p is out of the [0, 1] range (%v)", p)
	}
	if t < 0 || t > 1 {
		return 0, 0, 0, fmt.Errorf("t is out of the [0, 1] range (%v)", t)
	}

	prime := func(v float64) float64 {
		r := math.Pow(math.Abs(v), 1/0.43)
		if math.Signbit(v) {
			return -r
		}
		return r
	}

	lms := conversionIptLms.vdot(vector{i, p, t})
	xyz := conversionLmsXyz.vdot(lms.mapfunc(prime))

	return xyz.v0, xyz.v1, xyz.v2, nil
}

// SpectralToXYZ converts spectral readings to XYZ coordinates.
func SpectralToXYZ(color []float64, observer int, refIlluminant []float64) (x, y, z float64, err error) {
	var (
		stdObserverX = stdObs10X
		stdObserverY = stdObs10Y
		stdObserverZ = stdObs10Z
	)

	if observer == Observer2 {
		stdObserverX = stdObs2X
		stdObserverY = stdObs2Y
		stdObserverZ = stdObs2Z
	}

	l := len(color)
	if l != len(stdObserverX) || l != len(refIlluminant) {
		return 0, 0, 0, errors.New("mismatching spectral sampling length")
	}

	var (
		denom      float64 = 0
		xNumerator float64 = 0
		yNumerator float64 = 0
		zNumerator float64 = 0
	)
	for i := 0; i < l; i++ {
		denom += stdObserverY[i] * refIlluminant[i]

		sampleByRefIlluminant := color[i] * refIlluminant[i]
		xNumerator += sampleByRefIlluminant * stdObserverX[i]
		yNumerator += sampleByRefIlluminant * stdObserverY[i]
		zNumerator += sampleByRefIlluminant * stdObserverZ[i]
	}

	x = xNumerator / denom
	y = yNumerator / denom
	z = zNumerator / denom

	return x, y, z, nil
}

////////////////////////////////////////

// RGBtoCMYK converts a color from RGB coordinates to CMYK.
func RGBtoCMYK(r, g, b float64) (c, m, y, k float64, err error) {
	if c, m, y, err = RGBtoCMY(r, g, b); err != nil {
		return 0, 0, 0, 0, err
	}
	return CMYtoCMYK(c, m, y)
}

// CMYKtoRGB converts a color from CMYK coordinates to RGB.
func CMYKtoRGB(c, m, y, k float64) (r, g, b float64, err error) {
	if c, m, y, err = CMYKtoCMY(c, m, y, k); err != nil {
		return 0, 0, 0, err
	}
	return CMYtoRGB(c, m, y)
}

// RGBtoXYY converts a color from RGB coordinates to xyY.
func RGBtoXYY(r, g, b float64, space string) (float64, float64, float64, error) {
	if x, y, z, err := RGBtoXYZ(r, g, b, space); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoXYY(x, y, z)
	}
}

// XYYtoRGB converts a color from xyY coordinates to RGB.
func XYYtoRGB(x, y, Y float64, space string) (r, g, b float64, err error) {
	if xyzX, xyzY, xyzZ, err := XYYtoXYZ(x, y, Y); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoRGB(xyzX, xyzY, xyzZ, space)
	}
}

// RGBtoLAB converts a color from RGB coordinates to Lab.
func RGBtoLAB(r, g, b float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if x, y, z, err := RGBtoXYZ(r, g, b, space); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoLAB(x, y, z, observer, illuminant)
	}
}

// LABtoRGB converts a color from Lab coordinates to RGB.
func LABtoRGB(l, a, b float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if x, y, z, err := LABtoXYZ(l, a, b, observer, illuminant); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoRGB(x, y, z, space)
	}
}

// RGBtoLCHAB converts a color from RGB coordinates to LCHab.
func RGBtoLCHAB(r, g, b float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if l, a, b, err := RGBtoLAB(r, g, b, space, observer, illuminant); err != nil {
		return 0, 0, 0, err
	} else {
		return LABtoLCHAB(l, a, b)
	}
}

// LCHABtoRGB converts a color from LCHab coordinates to RGB.
func LCHABtoRGB(l, c, h float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if labL, labA, labB, err := LCHABtoLAB(l, c, h); err != nil {
		return 0, 0, 0, err
	} else {
		return LABtoRGB(labL, labA, labB, space, observer, illuminant)
	}
}

// RGBtoLUV converts a color from RGB coordinates to Luv.
func RGBtoLUV(r, g, b float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if x, y, z, err := RGBtoXYZ(r, g, b, space); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoLUV(x, y, z, observer, illuminant)
	}
}

// LUVtoRGB converts a color from Luv coordinates to RGB.
func LUVtoRGB(l, u, v float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if x, y, z, err := LUVtoXYZ(l, u, v, observer, illuminant); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoRGB(x, y, z, space)
	}
}

// RGBtoLCHUV converts a color from RGB coordinates to LCHuv.
func RGBtoLCHUV(r, g, b float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if l, u, v, err := RGBtoLUV(r, g, b, space, observer, illuminant); err != nil {
		return 0, 0, 0, err
	} else {
		return LUVtoLCHUV(l, u, v)
	}
}

// LCHUVtoRGB converts a color from LCHuv coordinates to RGB.
func LCHUVtoRGB(l, c, h float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if luvL, luvU, luvV, err := LCHUVtoLUV(l, c, h); err != nil {
		return 0, 0, 0, err
	} else {
		return LUVtoRGB(luvL, luvU, luvV, space, observer, illuminant)
	}
}

// RGBtoIPT converts a color from RGB coordinates to IPT.
func RGBtoIPT(r, g, b float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if x, y, z, err := RGBtoXYZ(r, g, b, space); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoIPT(x, y, z, observer, illuminant)
	}
}

// IPTtoRGB converts a color from IPT coordinates to RGB.
func IPTtoRGB(i, p, t float64, space string, observer int, illuminant string) (float64, float64, float64, error) {
	if x, y, z, err := IPTtoXYZ(i, p, t); err != nil {
		return 0, 0, 0, err
	} else {
		return XYZtoRGB(x, y, z, space)
	}
}

////////////////////////////////////////

func checkRGB(r, g, b float64) error {
	if r < 0 || r > 1 {
		return fmt.Errorf("red is out of the [0, 1] range (%v)", r)
	}
	if g < 0 || g > 1 {
		return fmt.Errorf("green is out of the [0, 1] range (%v)", g)
	}
	if b < 0 || b > 1 {
		return fmt.Errorf("blue is out of the [0, 1] range (%v)", b)
	}

	return nil
}

func checkYUV(y, u, v float64) error {
	if y < 0 || y > 1 {
		return fmt.Errorf("luma (Y) is out of the [0, 1] range (%v)", y)
	}
	if u < -0.5 || u > 0.5 {
		return fmt.Errorf("blue projection (U) is out of the [-0.5, 0.5] range (%v)", u)
	}
	if v < -0.5 || v > 0.5 {
		return fmt.Errorf("red projection (V) is out of the [-0.5, 0.5] range (%v)", v)
	}
	return nil
}

func checkXYZ(x, y, z float64) error {
	if x < 0 || x > 1 {
		return fmt.Errorf("x is out of the [0, 1] range (%v)", x)
	}
	if y < 0 || y > 1 {
		return fmt.Errorf("y is out of the [0, 1] range (%v)", y)
	}
	if z < 0 || z > 1 {
		return fmt.Errorf("z is out of the [0, 1] range (%v)", z)
	}
	return nil
}

func checkLAB(l, a, b float64) error {
	if l < 0 || l > 1 {
		return fmt.Errorf("L is out of the [0, 1] range (%v)", l)
	}
	if a < 0 || a > 1 {
		return fmt.Errorf("a is out of the [0, 1] range (%v)", l)
	}
	if b < 0 || b > 1 {
		return fmt.Errorf("b is out of the [0, 1] range (%v)", l)
	}
	return nil
}

func checkLUV(l, u, v float64) error {
	if l < 0 || l > 1 {
		return fmt.Errorf("L is out of the [0, 1] range (%v)", l)
	}
	if u < 0 || u > 1 {
		return fmt.Errorf("u is out of the [0, 1] range (%v)", u)
	}
	if v < 0 || v > 1 {
		return fmt.Errorf("v is out of the [0, 1] range (%v)", v)
	}
	return nil
}
