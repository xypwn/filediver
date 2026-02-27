package stingray

import "github.com/go-gl/mathgl/mgl32"

var ToGLTFMatrix mgl32.Mat4 = mgl32.Mat4([16]float32{
	1, 0, 0, 0,
	0, 0, -1, 0,
	0, 1, 0, 0,
	0, 0, 0, 1,
})

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
	return ToGLTFMatrix.Mul4x1(o.PositionVec.Vec4(1)).Vec3(), ToGLTFMatrix.Mul4x1(o.RotationVec), mgl32.Vec3{o.ScaleVec[0], o.ScaleVec[2], o.ScaleVec[1]}
}
