package disorder

type tag byte

const (
	tagUndefined tag = 0

	tagBool   tag = 1
	tagInt    tag = 2
	tagLong   tag = 3
	tagFloat  tag = 4
	tagDouble tag = 5
	tagBytes  tag = 6

	tagString    tag = 11
	tagTimestamp tag = 12
	tagEnum      tag = 13

	tagArrayStart  tag = 21
	tagArrayEnd    tag = 22
	tagObjectStart tag = 23
	tagObjectEnd   tag = 24
)
