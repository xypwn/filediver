package extractor

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type Object interface {
	Unit() stingray.Hash
	Position() mgl32.Vec3
	SetPosition(mgl32.Vec3)
	Rotation() mgl32.Vec4
	SetRotation(mgl32.Vec4)
	Scale() mgl32.Vec3
	SetScale(mgl32.Vec3)
	ToGLTF() (mgl32.Vec3, mgl32.Vec4, mgl32.Vec3)
}
