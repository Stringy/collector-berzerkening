extends RigidBody2D

enum ProjectileType {
	Process,
	Network,
	Endpoint
}

var process_sprite = preload("res://assets/sprites/ball_red.svg")
var network_sprite = preload("res://assets/sprites/ball_blue.svg")
var endpoint_sprite = preload("res://assets/sprites/ball_pink.svg")

signal exploded(child)

@export var kind: ProjectileType

# Called when the node enters the scene tree for the first time.
func _ready():
	match kind:
		ProjectileType.Process:
			$Sprite.texture = process_sprite
		ProjectileType.Network:
			$Sprite.texture = network_sprite
		ProjectileType.Endpoint:
			$Sprite.texture = endpoint_sprite

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	pass

func _on_visible_on_screen_enabler_2d_screen_exited():
	emit_signal("exploded", self)
	
func set_event_type(name):
	match name:
		"process":
			kind = ProjectileType.Process
		"connection":
			kind = ProjectileType.Network
		"endpoint":
			kind = ProjectileType.Endpoint
