package main

import (
	"testing"

	"github.com/normastars/frame"
)

func TestAdd(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "hello",
			args: args{
				a: 1,
				b: 2,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Add(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelloWorld(t *testing.T) {
	type args struct {
		c *frame.Context
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				c: frame.NewContextNoGin(`E:\src\go\workspace\frame\example\conf\default.json`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HelloWorld(tt.args.c)
		})
	}
}
