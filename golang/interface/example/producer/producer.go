package producer

import "io"

type Product struct {
}

type Producer interface {
	Produce() *Product
}

type defaultProducer struct {

}

func (p *defaultProducer) Produce() *Product {
	return nil
}


type DefaultProducer struct {
	reader io.Reader
}

func (p *DefaultProducer) Produce() *Product {
	return nil
}

func NewProducer () Producer {
	return &defaultProducer{}
}

func NewDefaultProducer() *DefaultProducer {
	return &DefaultProducer{}
}
