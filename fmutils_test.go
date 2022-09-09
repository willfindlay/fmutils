package fmutils

import (
	"fmt"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/mennanov/fmutils/testproto"
)

func Test_NestedMaskFromPaths(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want NestedMask
	}{
		{
			name: "no nested fields",
			args: args{paths: []string{"a", "b", "c"}},
			want: NestedMask{"a": NestedMask{}, "b": NestedMask{}, "c": NestedMask{}},
		},
		{
			name: "with nested fields",
			args: args{paths: []string{"aaa.bb.c", "dd.e", "f"}},
			want: NestedMask{
				"aaa": NestedMask{"bb": NestedMask{"c": NestedMask{}}},
				"dd":  NestedMask{"e": NestedMask{}},
				"f":   NestedMask{}},
		},
		{
			name: "single field",
			args: args{paths: []string{"a"}},
			want: NestedMask{"a": NestedMask{}},
		},
		{
			name: "empty fields",
			args: args{paths: []string{}},
			want: NestedMask{},
		},
		{
			name: "invalid input",
			args: args{paths: []string{".", "..", "..."}},
			want: NestedMask{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NestedMaskFromPaths(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NestedMaskFromPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createAny(m proto.Message) *anypb.Any {
	any, err := anypb.New(m)
	if err != nil {
		panic(err)
	}
	return any
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		msg   proto.Message
		want  proto.Message
	}{
		{
			name:  "empty mask keeps all the fields",
			paths: []string{},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
				LoginTimestamps: []int64{1, 2},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
				LoginTimestamps: []int64{1, 2},
			},
		},
		{
			name:  "mask with all root fields keeps all root fields",
			paths: []string{"user", "photo"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask with single root field keeps that field only",
			paths: []string{"user"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
			},
		},
		{
			name:  "mask with nested fields keeps the listed fields only",
			paths: []string{"user.name", "photo.path", "photo.dimensions.width"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					Name: "user name",
				},
				Photo: &testproto.Photo{
					Path: "photo path",
					Dimensions: &testproto.Dimensions{
						Width: 100,
					},
				},
			},
		},
		{
			name:  "mask with oneof field keeps the entire field",
			paths: []string{"user"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
		},
		{
			name:  "mask with nested oneof fields keeps listed fields only",
			paths: []string{"profile.photo.dimensions", "profile.user.user_id", "profile.login_timestamps"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
						Name:   "user name",
					},
					Photo: &testproto.Photo{
						PhotoId: 1,
						Path:    "photo path",
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
					},
					Photo: &testproto.Photo{
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
		},
		{
			name:  "mask with Any field in oneof field keeps the entire Any field",
			paths: []string{"details"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Details{Details: createAny(&testproto.Result{
					Data:      []byte("bytes"),
					NextToken: 1,
				})},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Details{Details: createAny(&testproto.Result{
					Data:      []byte("bytes"),
					NextToken: 1,
				})},
			},
		},
		{
			name:  "mask with repeated nested fields keeps the listed fields",
			paths: []string{"profile.gallery.photo_id", "profile.gallery.dimensions.height"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Dimensions: &testproto.Dimensions{
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Dimensions: &testproto.Dimensions{
									Height: 400,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with repeated field keeps the listed field only",
			paths: []string{"profile.gallery"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with map field keeps the listed field only",
			paths: []string{"profile.attributes.a1", "profile.attributes.a2.tags.t2", "profile.attributes.aNonExistant"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t2": "2",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Filter(tt.msg, tt.paths)
			if !proto.Equal(tt.msg, tt.want) {
				t.Errorf("msg %v, want %v", tt.msg, tt.want)
			}
		})
	}
}

func TestPrune(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		msg   proto.Message
		want  proto.Message
	}{
		{
			name:  "empty mask keeps all the fields",
			paths: []string{},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask all root fields clears all fields",
			paths: []string{"user", "photo"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{},
		},
		{
			name:  "mask with single root field clears that field only",
			paths: []string{"user"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask with nested fields clears that fields only",
			paths: []string{"user.name", "photo.path", "photo.dimensions.width"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Dimensions: &testproto.Dimensions{
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask with oneof field clears that entire field only",
			paths: []string{"user"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
			want: &testproto.Event{
				EventId: 1,
			},
		},
		{
			name:  "mask with nested oneof fields clears listed fields only",
			paths: []string{"profile.photo.dimensions", "profile.user.user_id", "profile.login_timestamps"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
						Name:   "user name",
					},
					Photo: &testproto.Photo{
						PhotoId: 1,
						Path:    "photo path",
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						Name: "user name",
					},
					Photo: &testproto.Photo{
						PhotoId: 1,
						Path:    "photo path",
					},
				}},
			},
		},
		{
			name:  "mask with repeated nested fields clears the listed fields",
			paths: []string{"profile.gallery.photo_id", "profile.gallery.dimensions.height"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								Path: "path 1",
								Dimensions: &testproto.Dimensions{
									Width: 100,
								},
							},
							{
								Path: "path 2",
								Dimensions: &testproto.Dimensions{
									Width: 300,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with repeated field clears the listed field only",
			paths: []string{"profile.gallery"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
					},
				},
			},
		},
		{
			name:  "mask with map field prunes the listed field",
			paths: []string{"profile.attributes.a1", "profile.attributes.a2.tags.t2", "profile.attributes.aNonExistant"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Prune(tt.msg, tt.paths)
			if !proto.Equal(tt.msg, tt.want) {
				t.Errorf("msg %v, want %v", tt.msg, tt.want)
			}
		})
	}
}

func BenchmarkNestedMaskFromPaths(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NestedMaskFromPaths([]string{"aaa.bbb.c.d.e.f", "aa.b.cc.ddddddd", "e", "f", "g.h.i.j.k"})
	}
}

func Test_WildcardWildcardNestedMaskFromPaths(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want WildcardNestedMask
	}{
		{
			name: "no nested fields",
			args: args{paths: []string{"a", "b", "c"}},
			want: WildcardNestedMask{"a": WildcardNestedMask{}, "b": WildcardNestedMask{}, "c": WildcardNestedMask{}},
		},
		{
			name: "with nested fields",
			args: args{paths: []string{"aaa.*.c", "dd.e", "f", "*.*.*.g"}},
			want: WildcardNestedMask{
				"aaa": WildcardNestedMask{"*": WildcardNestedMask{"c": WildcardNestedMask{}}},
				"dd":  WildcardNestedMask{"e": WildcardNestedMask{}},
				"f":   WildcardNestedMask{},
				"*":   WildcardNestedMask{"*": WildcardNestedMask{"*": WildcardNestedMask{"g": WildcardNestedMask{}}}}},
		},
		{
			name: "single field",
			args: args{paths: []string{"*"}},
			want: WildcardNestedMask{"*": WildcardNestedMask{}},
		},
		{
			name: "empty fields",
			args: args{paths: []string{}},
			want: WildcardNestedMask{},
		},
		{
			name: "invalid input",
			args: args{paths: []string{".", "..", "..."}},
			want: WildcardNestedMask{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WildcardNestedMaskFromPaths(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WildcardNestedMaskFromPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWildcardFilter(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		msg   proto.Message
		want  proto.Message
	}{
		{
			name:  "empty mask keeps all the fields",
			paths: []string{},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
				LoginTimestamps: []int64{1, 2},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
				LoginTimestamps: []int64{1, 2},
			},
		},
		{
			name:  "mask with all root fields keeps all root fields",
			paths: []string{"user", "photo"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask with single root field keeps that field only",
			paths: []string{"user"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
			},
		},
		{
			name:  "mask with nested fields keeps the listed fields only",
			paths: []string{"user.name", "photo.path", "photo.dimensions.width"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					Name: "user name",
				},
				Photo: &testproto.Photo{
					Path: "photo path",
					Dimensions: &testproto.Dimensions{
						Width: 100,
					},
				},
			},
		},
		{
			name:  "mask with oneof field keeps the entire field",
			paths: []string{"user"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
		},
		{
			name:  "mask with nested oneof fields keeps listed fields only",
			paths: []string{"profile.photo.dimensions", "profile.user.user_id", "profile.login_timestamps"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
						Name:   "user name",
					},
					Photo: &testproto.Photo{
						PhotoId: 1,
						Path:    "photo path",
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
					},
					Photo: &testproto.Photo{
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
		},
		{
			name:  "mask with Any field in oneof field keeps the entire Any field",
			paths: []string{"details"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Details{Details: createAny(&testproto.Result{
					Data:      []byte("bytes"),
					NextToken: 1,
				})},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Details{Details: createAny(&testproto.Result{
					Data:      []byte("bytes"),
					NextToken: 1,
				})},
			},
		},
		{
			name:  "mask with repeated nested fields keeps the listed fields",
			paths: []string{"profile.gallery.photo_id", "profile.gallery.dimensions.height"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Dimensions: &testproto.Dimensions{
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Dimensions: &testproto.Dimensions{
									Height: 400,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with repeated field keeps the listed field only",
			paths: []string{"profile.gallery"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with map field keeps the listed field only",
			paths: []string{"profile.attributes.a1", "profile.attributes.a2.tags.t2", "profile.attributes.aNonExistant"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t2": "2",
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with wildcard value",
			paths: []string{"profile.attributes.*", "profile.attributes.a2.tags.t2", "profile.attributes.aNonExistant"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t2": "2",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with lots of wildcard values",
			paths: []string{"profile.attributes.*.tags.t1", "profile.*.*.tags.t3"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t3": "3",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("new test", tt.name, "starts here -----------------")
			WildcardFilter(tt.msg, tt.paths)
			if !proto.Equal(tt.msg, tt.want) {
				t.Errorf("msg %v, want %v", tt.msg, tt.want)
			}
		})
	}
}

func BenchmarkWilcardNestedMaskFromPaths(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WildcardNestedMaskFromPaths([]string{"aaa.bbb.c.d.e.f", "aa.b.cc.ddddddd", "e", "f", "g.h.i.j.k", "*.*.*.*.a.*.*.bbbbb.*.cccccc"})
	}
}
