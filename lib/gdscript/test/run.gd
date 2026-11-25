# Test suite for disorder.gd. Run from lib/gdscript:
#   godot --headless --script test/run.gd [-- <dir>]
#
# With <dir>, it also reads object.bin / object.json written by the Go tests,
# and writes from_gd.bin / from_gd.json for Go to read back.
extends SceneTree

const Codec = preload("../disorder.gd")
const Schema = preload("generated/test_data/test/schema.gd")
const SubSchema = preload("generated/test_data/test/sub/sub_schema.gd")

var failures := 0
var checks := 0


func _init() -> void:
	test_round_trip()
	test_json_round_trip()
	test_deterministic_and_sorted()
	test_unknown_fields_are_skipped()
	test_errors()
	test_timestamps()
	var args := OS.get_cmdline_user_args()
	if args.size() > 0:
		test_interop(args[0])
	print("%d checks, %d failures" % [checks, failures])
	quit(1 if failures > 0 else 0)


func check(condition: bool, message: String) -> void:
	checks += 1
	if not condition:
		failures += 1
		printerr("FAIL: " + message)


func check_eq(actual: Variant, expected: Variant, message: String) -> void:
	checks += 1
	if typeof(actual) != typeof(expected) or actual != expected:
		failures += 1
		printerr("FAIL: %s: got %s, want %s" % [message, str(actual), str(expected)])


func number_wrapper(value: int) -> Object:
	var wrapper = SubSchema.DisNumberWrapper.new()
	wrapper.value = SubSchema.DisNumber.new()
	wrapper.value.value = value
	return wrapper


# Same values as newObject() in internal/disorder_test.go.
func new_object() -> Object:
	var o = Schema.DisObject.new()
	o.bool_field = true
	o.int_field = 123
	o.long_field = 1 << 40
	o.float_field = 1.5
	o.double_field = -2.25
	o.string_field = "foo"
	o.bytes_fields = PackedByteArray([0, 1, 2, 255])
	o.enum_field = Schema.DisColor.BLUE
	o.time_field = 1700000000123
	o.obj_field = number_wrapper(789)
	o.int_array.assign([1, 2, 3])
	o.int_map = {"4": 4, "5": 5, "6": 6}
	o.obj_array.assign([number_wrapper(1), number_wrapper(2)])
	o.obj_map = {"789": number_wrapper(789)}
	o.nested = {"key0": {"key1": [[{"key2": Schema.DisColor.BLUE}]]}}
	return o


func check_object(o: Object, label: String) -> void:
	if o == null:
		check(false, "%s: decode failed: %s" % [label, Codec.last_error])
		return
	check_eq(o.bool_field, true, label + " bool")
	check_eq(o.int_field, 123, label + " int")
	check_eq(o.long_field, 1 << 40, label + " long")
	check_eq(o.float_field, 1.5, label + " float")
	check_eq(o.double_field, -2.25, label + " double")
	check_eq(o.string_field, "foo", label + " string")
	check_eq(o.bytes_fields, PackedByteArray([0, 1, 2, 255]), label + " bytes")
	check_eq(o.enum_field, Schema.DisColor.BLUE, label + " enum")
	check_eq(o.time_field, 1700000000123, label + " time")
	check_eq(o.obj_field.value.value, 789, label + " obj")
	check_eq(o.int_array, [1, 2, 3] as Array[int], label + " int_array")
	check_eq(o.int_map, {"4": 4, "5": 5, "6": 6}, label + " int_map")
	check_eq(o.obj_array.size(), 2, label + " obj_array size")
	check_eq(o.obj_array[1].value.value, 2, label + " obj_array[1]")
	check_eq(o.obj_map["789"].value.value, 789, label + " obj_map")
	check_eq(o.nested["key0"]["key1"][0][0]["key2"], Schema.DisColor.BLUE, label + " nested")
	check(o.empty_obj == null, label + " empty_obj stays null")
	check_eq(o.empty_string, "", label + " empty_string")


func test_round_trip() -> void:
	var bytes := Codec.encode(new_object())
	check_eq(Codec.last_error, "", "encode error")
	check(bytes.size() > 0, "encode produced data")
	check_object(Codec.decode(bytes, Schema.DisObject), "binary")


func test_json_round_trip() -> void:
	var json := Codec.to_json(new_object())
	check_eq(Codec.last_error, "", "to_json error")
	var parsed = JSON.parse_string(json)
	check_eq(parsed["bytes_fields"], "AAEC/w==", "bytes are base64 in json")
	check_eq(parsed["enum_field"], "blue", "enum is a string in json")
	check_eq(parsed["time_field"], "2023-11-14T22:13:20.123Z", "timestamp is rfc3339 in json")
	check(not parsed.has("empty_obj") and not parsed.has("empty_array"), "empty values are omitted")
	check_object(Codec.from_json(json, Schema.DisObject), "json")


func test_deterministic_and_sorted() -> void:
	var a := Codec.encode(new_object())
	var b := Codec.encode(new_object())
	check_eq(a, b, "encoding is deterministic")
	var o = Schema.DisObject.new()
	o.int_map = {"b": 1, "a": 2, "c": 3}
	var decoded = Codec.decode(Codec.encode(o), Schema.DisObject)
	check_eq(decoded.int_map.keys(), ["a", "b", "c"], "map keys are written sorted")


func test_unknown_fields_are_skipped() -> void:
	# Decode a Number into Zero: field "value" is unknown there.
	var number = SubSchema.DisNumber.new()
	number.value = 5
	var decoded = Codec.decode(Codec.encode(number), Schema.DisZero)
	check(decoded != null, "unknown fields are skipped: " + Codec.last_error)
	check_eq(decoded.zero_array.size(), 0, "unknown fields leave defaults")
	# Skip containers and every scalar kind.
	var full = Codec.decode(Codec.encode(new_object()), Schema.DisZero)
	check(full != null, "skipping a full object works: " + Codec.last_error)


func test_errors() -> void:
	check_eq(Codec.encode(null).size(), 0, "encode null")
	check(Codec.last_error != "", "encode null sets error")

	var bytes := Codec.encode(new_object())
	check(Codec.decode(bytes.slice(0, bytes.size() - 1), Schema.DisObject) == null, "truncated data")
	check(Codec.decode(PackedByteArray(), Schema.DisObject) == null, "empty data")
	check(Codec.decode(Codec.encode(new_object()), SubSchema.DisNumber) != null, "other class skips unknown fields")

	# Type mismatch: a Number has field "value" as int, Wrapper expects an object.
	var number = SubSchema.DisNumber.new()
	number.value = 1
	check(Codec.decode(Codec.encode(number), SubSchema.DisNumberWrapper) == null, "type mismatch is an error")
	check(Codec.last_error.contains("mismatch"), "mismatch message: " + Codec.last_error)

	var o = Schema.DisObject.new()
	o.enum_field = 99
	check_eq(Codec.encode(o).size(), 0, "invalid enum value")
	o = Schema.DisObject.new()
	o.int_map = {1: 1}
	check_eq(Codec.encode(o).size(), 0, "non-string map key")

	check(Codec.from_json("{", Schema.DisObject) == null, "invalid json")
	check(Codec.from_json("{\"int_field\": \"x\"}", Schema.DisObject) == null, "json type mismatch")
	check(Codec.from_json("{\"enum_field\": \"purple\"}", Schema.DisObject) == null, "json invalid enum")
	check(Codec.from_json("{\"time_field\": \"yesterday\"}", Schema.DisObject) == null, "json invalid timestamp")
	var ok = Codec.from_json("{\"unknown\": 1}", Schema.DisObject)
	check(ok != null, "json unknown fields are ignored")


func test_timestamps() -> void:
	var o = Schema.DisObject.new()
	for text in ["2023-11-14T22:13:20.123Z", "2023-11-14T23:13:20.123+01:00", "2023-11-14T17:13:20.123-05:00", "2023-11-14T22:13:20.123456789Z"]:
		var decoded = Codec.from_json(JSON.stringify({"time_field": text}), Schema.DisObject)
		check(decoded != null, "parse " + text + ": " + Codec.last_error)
		if decoded != null:
			check_eq(decoded.time_field, 1700000000123, "timestamp " + text)
	o.time_field = -1500
	var round_trip = Codec.from_json(Codec.to_json(o), Schema.DisObject)
	check_eq(round_trip.time_field, -1500, "negative timestamp")


func test_interop(dir: String) -> void:
	var bin := FileAccess.get_file_as_bytes(dir.path_join("object.bin"))
	check(bin.size() > 0, "read object.bin")
	var from_go = Codec.decode(bin, Schema.DisObject)
	check_object(from_go, "go binary")
	var json := FileAccess.get_file_as_string(dir.path_join("object.json"))
	check_object(Codec.from_json(json, Schema.DisObject), "go json")

	var out := FileAccess.open(dir.path_join("from_gd.bin"), FileAccess.WRITE)
	out.store_buffer(Codec.encode(new_object()))
	out.close()
	out = FileAccess.open(dir.path_join("from_gd.json"), FileAccess.WRITE)
	out.store_string(Codec.to_json(new_object()))
	out.close()
