package unit

import (
    "testing"
    "github.com/egocentri/finalproject/internal/services"
)

func TestEvaluateBasic(t *testing.T) {
    v, err := services.Evaluate("2+3*4")
    if err != nil {
        t.Fatal(err)
    }
    if v != 14 {
        t.Fatalf("expected 14, got %v", v)
    }
}

func TestEvaluateParen(t *testing.T) {
    v, err := services.Evaluate("(1+2)*3")
    if err != nil {
        t.Fatal(err)
    }
    if v != 9 {
        t.Fatalf("expected 9, got %v", v)
    }
}
