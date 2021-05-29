//-----------------------------------------------------------------------------
/*

Pipe hose for diaphragm pump

*/
//-----------------------------------------------------------------------------

package main

import (
	"fmt"
	"log"

	"github.com/deadsy/sdfx/obj"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
)

//-----------------------------------------------------------------------------

var thread_diameter = 44.0
var thread_pitch = 3.0
var cap_radius = thread_diameter/2.0 + 5
var cap_height = cap_thickness + 15.0
var cap_thickness = 5.0

var ri = 19.0 / 2
var ra = ri + 1.5
var h = 5.0
var numRing = 7
var rhole = 25 / 2.0

// var rhole = ri - 2.0

//-----------------------------------------------------------------------------

func cap_outer() sdf.SDF3 {
	t, err := obj.KnurledHead3D(cap_radius, cap_height, cap_radius*0.25)
	if err != nil {
		log.Panic(err)
	}
	return t
}

func cap_inner() sdf.SDF3 {
	tp, err := sdf.PlasticButtressThread(thread_diameter/2.0, thread_pitch)
	if err != nil {
		log.Panic(err)
	}
	screw, err := sdf.Screw3D(tp, cap_height, thread_pitch, 1)
	if err != nil {
		log.Panic(err)
	}
	return sdf.Transform3D(screw, sdf.Translate3d(sdf.V3{0, 0, -cap_thickness}))
}

func full_cap() sdf.SDF3 {
	return sdf.Difference3D(
		sdf.Union3D(
			cap_outer(),
			tulle(),
		),
		sdf.Union3D(
			cap_inner(),
			hole()),
	)
}
func tulle_inlet() sdf.SDF3 {

	t, err := sdf.Cylinder3D(10*(cap_height), ri-1.2, 0)
	if err != nil {
		log.Panic(err)
	}
	return sdf.Difference3D(
		sdf.Union3D(
			inlet(),
			tulle(),
		),
		sdf.Union3D(

			t),
	)
}

func ring_cap() sdf.SDF3 {
	return sdf.Difference3D(
		sdf.Union3D(
			cap_outer(),
			// tulle(),
		),
		sdf.Union3D(
			cap_inner(),
			hole()),
	)
}

func hole() sdf.SDF3 {
	t, err := sdf.Cylinder3D(2*(cap_height+float64(numRing)*h), rhole, 0)
	if err != nil {
		log.Panic(err)
	}
	return t
}

func tulle() sdf.SDF3 {

	points := []sdf.V2{
		{0, 0},
		{ri, 0},
		{ra, h / 3},
		{ri, h},
		{0, h},
	}

	var rings []sdf.SDF2
	for i := 0; i < numRing; i++ {
		s0, err := sdf.Polygon2D(points)
		if err != nil {
			log.Panic(err)
		}
		s0 = sdf.Transform2D(s0, sdf.Translate2d(sdf.V2{0, (h * float64(i)) + cap_height/2}))
		rings = append(rings, s0)
	}
	s := sdf.Union2D(rings...)
	s1, err := sdf.Revolve3D(s)
	if err != nil {
		log.Panic(err)
	}
	return s1
}
func inlet() sdf.SDF3 {
	rU := 29.0 / 2
	rM := 30.0 / 2
	rO := 36.0 / 2
	hU := 10.0
	hO := 3.0
	points := []sdf.V2{
		{0, 0},
		{rU, 0},
		{rM, hU},
		{rO, hU},
		{rO, hO + hU},
		{0, hO + hU},
	}

	s0, err := sdf.Polygon2D(points)
	if err != nil {
		log.Panic(err)
	}
	s1, err := sdf.Revolve3D(s0)
	if err != nil {
		log.Panic(err)
	}
	return s1
}

//---------------------------------go--------------------------------------------

func main() {
	// render.RenderSTLSlow(ring_cap(), 300, fmt.Sprintf("ring_cap_%.0fx%.0f.stl", thread_diameter, thread_pitch))
	render.RenderSTLSlow(tulle_inlet(), 250, fmt.Sprintf("tulle_inlet%.0fx%.0f.stl", thread_diameter, thread_pitch))
	// render.RenderSTLSlow(full_cap(), 200, "sack_adapter.stl")

	// render.RenderSTLSlow(full_cap(), 300, "cap.stl")
}

//-----------------------------------------------------------------------------
