package classifier

type Segments []Segment

func (ss *Segments) IsValid() (valid bool) {
	if len(*ss) == 0 {
		goto end
	}
	for _, segment := range *ss {
		if segment.IsValid() {
			continue
		}
		goto end
	}
end:
	return valid
}

type Segment struct {
	Type    SegmentType
	TagType TagType
}

func (s *Segment) IsValidTagType() bool {
	return s.TagType.IsValid()
}

func (s *Segment) IsValid() (valid bool) {
	if s.Type == InvalidSegmentType {
		goto end
	}
	if s.Type == CompleteTag && s.TagType == NotApplicable {
		goto end
	}
	valid = true
end:
	return valid
}

func (s *Segment) WithType(typ SegmentType) (segment Segment) {
	segment = *s
	segment.Type = typ
	return segment
}

func (s *Segment) Invalidate() {
	s.Type = InvalidSegmentType
}

func (s *Segments) Reset() {
	if len(*s) == 0 {
		goto end
	}
	(*s)[0] = Segment{}
end:
	return
}
