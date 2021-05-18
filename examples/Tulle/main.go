//-----------------------------------------------------------------------------
/*

Pipe hose for diaphragm pump

*/
//-----------------------------------------------------------------------------

package main

import (
	"log"

	"github.com/deadsy/sdfx/obj"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
)

//-----------------------------------------------------------------------------

var cap_radius = thread_radius + 6
var cap_height = 23.0
var cap_thickness = 8.0
var thread_pitch = 1.5

var thread_diameter = 18.5
var thread_radius = thread_diameter / 2.0

//-----------------------------------------------------------------------------

func cap_outer() sdf.SDF3 {
	t, err := obj.KnurledHead3D(cap_radius, cap_height, cap_radius*0.25)
	if err != nil {
		log.Panic(err)
	}
	return t
}

func cap_inner() sdf.SDF3 {
	tp, err := sdf.ISOThread(thread_radius, thread_pitch, true)
	if err != nil {
		log.Panic(err)
	}
	screw, err := sdf.Screw3D(tp, cap_height, thread_pitch, 1)
	if err != nil {
		log.Panic(err)
	}
	return sdf.Transform3D(screw, sdf.Translate3d(sdf.V3{0, 0, -cap_thickness}))
}

func gas_cap() sdf.SDF3 {
	return sdf.Difference3D(cap_outer(), sdf.Union3D(cap_inner(), hole()))
}

func hole() sdf.SDF3 {
	t, err := sdf.Cylinder3D(cap_height, 12.0/2.0, 0)
	if err != nil {
		log.Panic(err)
	}
	return t
}

//---------------------------------go--------------------------------------------

func main() {
	render.RenderSTLSlow(gas_cap(), 200, "cap.stl")
	// render.RenderSTLSlow(gas_cap(), 300, "cap.stl")
}

//-----------------------------------------------------------------------------
