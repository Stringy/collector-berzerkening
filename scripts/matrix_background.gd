extends Node2D

@export var font: Font
@export var width: int = 73
@export var height: int = 40

var col
var matrix = []
var upd = 0
var move = 0

# Called when the node enters the scene tree for the first time.
func _ready():
	randomize()
	for x in range(width):
		matrix.append([])
		for y in range(height):
			if y == 0:
				matrix[x].append(Vector2(randi() & 1, int(randi() % 11)))
			else:
				matrix[x].append(Vector2(0, 1))
	set_physics_process(true)

# Called every frame. 'delta' is the elapsed time since the previous frame.
func do_matrix():
	if upd == 10:
		upd = 0
	for x in range(width):
		matrix[x][0] = Vector2(randi_range(0, 9), randi()%11)
	
	for x in range(width):
		if matrix[x][0].y > 0:
			matrix[x].push_front(Vector2(matrix[x][0].x, matrix[x][0].y - 1))
			matrix[x].pop_back()
		elif matrix[x][upd].y <= 0:
			matrix[x].push_front(Vector2(randi()&1, randi()%11))
			matrix[x].pop_back()
	upd += 1
	
func _physics_process(delta):
	# any faster and it gets a bit queasy
	move += delta
	if move > 0.07:
		do_matrix()
		move = 0
		queue_redraw()
		
func _draw():
	for x in range(width):
		for y in range(height):
			if matrix[x][y].y == 10:
				col = Color(1, 1, 1, matrix[x][y].y * 0.1)
			else:
				col = Color(0, 1, 0, matrix[x][y].y * 0.1)
			draw_string(font, Vector2(x * 16, y * 16 + 16), str(matrix[x][y].x), HORIZONTAL_ALIGNMENT_LEFT, -1, 16, col)
		
