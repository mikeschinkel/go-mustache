package classifier

import (
	"testing"
)

func TestClassifierWithEnclosingTags(t *testing.T) {
	testsRunner(t, []TestCase{
		{
			name:     "Inverted Section - With content",
			template: "{{^section}}\n  Content\n{{/section}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginInvertedSegments},
				1: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				2: {Type: StandaloneLine, Segments: sectionEndInvertedSegments},
			},
			comment: "Inverted section with content should have a standalone opening and closing line",
		},
		{
			name:     "Inverted Section - With content and trailing newline",
			template: "{{^section}}\n  Content\n{{/section}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginInvertedSegments},
				1: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				2: {Type: StandaloneLine, Segments: sectionEndInvertedSegments},
				3: {Type: EmptyLine},
			},
			comment: "Inverted section with content should have a standalone opening and closing line and trailing newline",
		},
	})
}
