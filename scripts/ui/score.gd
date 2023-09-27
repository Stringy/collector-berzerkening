extends Label

@export var label: String = "Label"
var score = 0

func add(val: int = 1):
	score += val
	_update_text()

# Called when the node enters the scene tree for the first time.
func _ready():
	_update_text()
	
# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	pass

func _update_text():
	self.text = label + ": " + str(score)
