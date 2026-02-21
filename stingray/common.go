package stingray

import "github.com/go-gl/mathgl/mgl32"

type Transform struct {
	PositionVec mgl32.Vec3 `json:"position"`
	RotationVec mgl32.Vec4 `json:"rotation"`
	ScaleVec    mgl32.Vec3 `json:"scale"`
}

func (o *Transform) Position() mgl32.Vec3 {
	return o.PositionVec
}

func (o *Transform) Rotation() mgl32.Vec4 {
	return o.RotationVec
}

func (o *Transform) Scale() mgl32.Vec3 {
	return o.ScaleVec
}

func (o *Transform) SetPosition(v mgl32.Vec3) {
	o.PositionVec = v
}

func (o *Transform) SetRotation(v mgl32.Vec4) {
	o.RotationVec = v
}

func (o *Transform) SetScale(v mgl32.Vec3) {
	o.ScaleVec = v
}

func (o *Transform) ToGLTF() (mgl32.Vec3, mgl32.Vec4, mgl32.Vec3) {
	// // Convert to glTF coords
	// p := o.PositionVec[:]
	// p[1], p[2] = p[2], -p[1]

	// r := o.RotationVec[:]
	// r[1], r[2] = r[2], -r[1]

	// s := o.ScaleVec[:]
	gltfMatrix := mgl32.Mat4([16]float32{
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, -1, 0, 0,
		0, 0, 0, 1,
	})
	// return gltfMatrix.Mul4x1(o.PositionVec.Vec4(1)).Vec3(), o.RotationVec, o.ScaleVec
	rotationVec := gltfMatrix.Mul4x1(o.RotationVec)
	return gltfMatrix.Mul4x1(o.PositionVec.Vec4(1)).Vec3(), rotationVec, o.ScaleVec
}
