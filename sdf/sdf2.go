//-----------------------------------------------------------------------------
/*

2D Signed Distance Functions

*/
//-----------------------------------------------------------------------------

package sdf

import "math"

//-----------------------------------------------------------------------------

type SDF2 interface {
	Evaluate(p V2) float64
	BoundingBox() Box2
}

//-----------------------------------------------------------------------------
// Basic SDF Functions

/*
func sdf_box2d(p, s V2) float64 {
	d := p.Abs().Sub(s)
	return d.Max(V2{0, 0}).Length() + Min(d.MaxComponent(), 0)
}
*/

func sdf_box2d(p, s V2) float64 {
	p = p.Abs()
	d := p.Sub(s)
	k := s.Y - s.X
	if d.X > 0 && d.Y > 0 {
		return d.Length()
	}
	if p.Y-p.X > k {
		return d.Y
	}
	return d.X
}

//-----------------------------------------------------------------------------
// 2D Circle

type CircleSDF2 struct {
	radius float64
	bb     Box2
}

func NewCircleSDF2(radius float64) SDF2 {
	s := CircleSDF2{}
	s.radius = radius
	d := V2{radius, radius}
	s.bb = Box2{d.Negate(), d}
	return &s
}

func (s *CircleSDF2) Evaluate(p V2) float64 {
	return p.Length() - s.radius
}

func (s *CircleSDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------
// 2D Box (rounded corners with radius > 0)

type BoxSDF2 struct {
	size   V2
	radius float64
	bb     Box2
}

func NewBoxSDF2(size V2, radius float64) SDF2 {
	size = size.MulScalar(0.5)
	s := BoxSDF2{}
	s.size = size.SubScalar(radius)
	s.radius = radius
	s.bb = Box2{size.Negate(), size}
	return &s
}

func (s *BoxSDF2) Evaluate(p V2) float64 {
	return sdf_box2d(p, s.size) - s.radius
}

func (s *BoxSDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------
// 2D Line, infinite line used for intersections and differences

type LineSDF2 struct {
	a      V2 // point on line
	normal V2 // normal to line, to the right of the line direction
}

// Return an infinite 2D line passing through a in direction v.
func NewLineSDF2(a, v V2) SDF2 {
	s := LineSDF2{}
	s.a = a
	s.normal = V2{v.Y, -v.X}.Normalize()
	return &s
}

// Return the minimum distance to the line (right side >0, left side < 0).
func (s *LineSDF2) Evaluate(p V2) float64 {
	return p.Sub(s.a).Dot(s.normal)
}

// Return the bounding box for the line (zero size).
func (s *LineSDF2) BoundingBox() Box2 {
	return Box2{}
}

//-----------------------------------------------------------------------------
// Offset an SDF2 - add a constant to the distance function
// offset > 0, enlarges and adds rounding to convex corners of the SDF
// offset < 0, skeletonizes the SDF

type OffsetSDF2 struct {
	sdf    SDF2
	offset float64
	bb     Box2
}

func NewOffsetSDF2(sdf SDF2, offset float64) SDF2 {
	s := OffsetSDF2{}
	s.sdf = sdf
	s.offset = offset
	// work out the bounding box
	bb := sdf.BoundingBox()
	s.bb = NewBox2(bb.Center(), bb.Size().AddScalar(2*offset))
	return &s
}

func (s *OffsetSDF2) Evaluate(p V2) float64 {
	return s.sdf.Evaluate(p) - s.offset
}

func (s *OffsetSDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------
// 2D Polygon

type PolySDF2 struct {
	line []Line2 // line segments
	bb   Box2    // bounding box
}

func NewPolySDF2(points []V2) SDF2 {
	n := len(points)
	if n < 3 {
		return nil
	}
	// build the line segments
	s := PolySDF2{}
	for i := 0; i < n-1; i++ {
		s.line = append(s.line, NewLine2(points[i], points[i+1]))
	}
	// close the loop
	s.line = append(s.line, NewLine2(points[n-1], points[0]))
	// build the bounding box
	l := s.line[0]
	a := l.A.Min(l.B)
	b := l.A.Max(l.B)
	for _, l := range s.line {
		a = a.Min(l.A.Min(l.B))
		b = b.Max(l.A.Max(l.B))
	}
	s.bb = Box2{a, b}
	return &s
}

func (s *PolySDF2) Evaluate(p V2) float64 {
	dd := math.MaxFloat64 // d^2 to polygon (>0)
	wn := 0               // winding number (inside/outside)
	// iterate over the line segments
	for _, l := range s.line {
		// record the minimum distance squared
		x := l.Distance2(p)
		if Abs(x) < dd {
			dd = Abs(x)
		}

		// Is the point in the polygon?
		// See: http://geomalgorithms.com/a03-_inclusion.html
		if l.A.Y <= p.Y {
			if l.B.Y > p.Y { // upward crossing
				if x < 0 { // p is to the left of the line segment
					wn += 1 // up intersect
				}
			}
		} else {
			if l.B.Y <= p.Y { // downward crossing
				if x > 0 { // p is to the right of the line segment
					wn -= 1 // down intersect
				}
			}
		}
	}
	// normalise d*d to d
	d := math.Sqrt(dd)
	if wn != 0 {
		// p is inside the polygon
		return -d
	}
	// p is outside the polygon
	return d
}

func (s *PolySDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------
// Transform SDF2

type TransformSDF2 struct {
	sdf     SDF2
	matrix  M33
	inverse M33
	bb      Box2
}

func NewTransformSDF2(sdf SDF2, matrix M33) SDF2 {
	s := TransformSDF2{}
	s.sdf = sdf
	s.matrix = matrix
	s.inverse = matrix.Inverse()
	s.bb = matrix.MulBox(sdf.BoundingBox())
	return &s
}

func (s *TransformSDF2) Evaluate(p V2) float64 {
	q := s.inverse.MulPosition(p)
	return s.sdf.Evaluate(q)
}

func (s *TransformSDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------
// ArraySDF2: Create an X by Y array of a given SDF2
// num = the array size
// size = the step size

type ArraySDF2 struct {
	sdf  SDF2
	num  V2i
	step V2
	min  MinFunc
	k    float64
	bb   Box2
}

func NewArraySDF2(sdf SDF2, num V2i, step V2) SDF2 {
	// check the number of steps
	if num[0] <= 0 || num[1] <= 0 {
		return nil
	}
	s := ArraySDF2{}
	s.sdf = sdf
	s.num = num
	s.step = step
	s.min = NormalMin
	// work out the bounding box
	bb0 := sdf.BoundingBox()
	bb1 := bb0.Translate(step.Mul(num.SubScalar(1).ToV2()))
	s.bb = bb0.Extend(bb1)
	return &s
}

// set the minimum function to control blending
func (s *ArraySDF2) SetMin(min MinFunc, k float64) {
	s.min = min
	s.k = k
}

func (s *ArraySDF2) Evaluate(p V2) float64 {
	d := math.MaxFloat64
	for j := 0; j < s.num[0]; j++ {
		for k := 0; k < s.num[1]; k++ {
			x := p.Sub(V2{float64(j) * s.step.X, float64(k) * s.step.Y})
			d = s.min(d, s.sdf.Evaluate(x), s.k)
		}
	}
	return d
}

func (s *ArraySDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------

type RotateSDF2 struct {
	sdf  SDF2
	num  int
	step M33
	min  MinFunc
	k    float64
	bb   Box2
}

func NewRotateSDF2(sdf SDF2, num int, step M33) SDF2 {
	// check the number of steps
	if num <= 0 {
		return nil
	}
	s := RotateSDF2{}
	s.sdf = sdf
	s.num = num
	s.step = step.Inverse()
	s.min = NormalMin
	// work out the bounding box
	v := sdf.BoundingBox().Vertices()
	bb_min := v[0]
	bb_max := v[0]
	for i := 0; i < s.num; i++ {
		bb_min = bb_min.Min(v.Min())
		bb_max = bb_max.Max(v.Max())
		v.MulVertices(step)
	}
	s.bb = Box2{bb_min, bb_max}
	return &s
}

// Return the minimum distance to the object.
func (s *RotateSDF2) Evaluate(p V2) float64 {
	d := math.MaxFloat64
	rot := Identity2d()
	for i := 0; i < s.num; i++ {
		x := rot.MulPosition(p)
		d = s.min(d, s.sdf.Evaluate(x), s.k)
		rot = rot.Mul(s.step)
	}
	return d
}

// Set the minimum function to control blending.
func (s *RotateSDF2) SetMin(min MinFunc, k float64) {
	s.min = min
	s.k = k
}

// Return the bounding box.
func (s *RotateSDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------

type UnionSDF2 struct {
	s0  SDF2
	s1  SDF2
	min MinFunc
	k   float64
	bb  Box2
}

func NewUnionSDF2(s0, s1 SDF2) SDF2 {
	if s0 == nil {
		return s1
	}
	if s1 == nil {
		return s0
	}
	s := UnionSDF2{}
	s.s0 = s0
	s.s1 = s1
	s.min = NormalMin
	s.bb = s0.BoundingBox().Extend(s1.BoundingBox())
	return &s
}

// set the minimum function to control blending
func (s *UnionSDF2) SetMin(min MinFunc, k float64) {
	s.min = min
	s.k = k
}

func (s *UnionSDF2) Evaluate(p V2) float64 {
	a := s.s0.Evaluate(p)
	b := s.s1.Evaluate(p)
	return s.min(a, b, s.k)
}

func (s *UnionSDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------
