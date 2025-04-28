package services

import (
    "errors"
    "strconv"
    "strings"
)

type parser struct {
    expr    string
    pos     int
    current rune
}

func (p *parser) next() {
    p.pos++
    if p.pos < len(p.expr) {
        p.current = rune(p.expr[p.pos])
    } else {
        p.current = 0
    }
}

func (p *parser) parse() (float64, error) {
    p.next()
    res, err := p.parseExpression()
    if err != nil {
        return 0, err
    }
    if p.current != 0 {
        return 0, errors.New("unexpected character")
    }
    return res, nil
}

func (p *parser) parseExpression() (float64, error) {
    acc, err := p.parseTerm()
    if err != nil {
        return 0, err
    }
    for p.current == '+' || p.current == '-' {
        op := p.current
        p.next()
        rhs, err := p.parseTerm()
        if err != nil {
            return 0, err
        }
        if op == '+' {
            acc += rhs
        } else {
            acc -= rhs
        }
    }
    return acc, nil
}

func (p *parser) parseTerm() (float64, error) {
    acc, err := p.parseFactor()
    if err != nil {
        return 0, err
    }
    for p.current == '*' || p.current == '/' {
        op := p.current
        p.next()
        rhs, err := p.parseFactor()
        if err != nil {
            return 0, err
        }
        if op == '*' {
            acc *= rhs
        } else {
            if rhs == 0 {
                return 0, errors.New("division by zero")
            }
            acc /= rhs
        }
    }
    return acc, nil
}

func (p *parser) parseFactor() (float64, error) {
    if p.current == '(' {
        p.next()
        res, err := p.parseExpression()
        if err != nil {
            return 0, err
        }
        if p.current != ')' {
            return 0, errors.New("mismatched parentheses")
        }
        p.next()
        return res, nil
    }
    start := p.pos
    for (p.current >= '0' && p.current <= '9') || p.current == '.' {
        p.next()
    }
    if start == p.pos {
        return 0, errors.New("expected number")
    }
    return strconv.ParseFloat(p.expr[start:p.pos], 64)
}

// Evaluate — внешняя точка входа
func Evaluate(expression string) (float64, error) {
    cleaned := strings.ReplaceAll(expression, " ", "")
    p := &parser{expr: cleaned, pos: -1}
    return p.parse()
}
