package flowl

import (
	"errors"
	"path"
	"strings"

	"github.com/cofunclabs/cofunc/internal/functiondriver"
)

// LoadedLocation
//
type LoadedLocation struct {
	DriverName   string
	FunctionName string
	Path         string
}

// FunctionNode
//
type FunctionNode struct {
	Name     string
	Driver   functiondriver.FunctionDriver
	Parallel *FunctionNode
	Args     map[string]string
}

// RunQueue
//
type RunQueue struct {
	Locations       map[string]LoadedLocation
	ConfiguredNodes map[string]*FunctionNode
	Queue           []*FunctionNode
}

func NewRunQueue() *RunQueue {
	return &RunQueue{
		Locations:       make(map[string]LoadedLocation),
		ConfiguredNodes: make(map[string]*FunctionNode),
		Queue:           make([]*FunctionNode, 0),
	}
}

func (rq *RunQueue) Generate(bs *BlockStore) error {
	if err := rq.processLoad(bs); err != nil {
		return err
	}
	if err := rq.processFn(bs); err != nil {
		return err
	}
	if err := rq.processRun(bs); err != nil {
		return err
	}
	return nil
}

func (rq *RunQueue) createFunctionNode(nodeName, fName string) (*FunctionNode, error) {
	loc, ok := rq.Locations[fName]
	if !ok {
		return nil, errors.New("not load function: " + fName)
	}
	loadTarget := loc.DriverName + ":" + loc.Path
	driver := functiondriver.New(loadTarget)
	if driver == nil {
		return nil, errors.New("not found driver: " + loadTarget)
	}
	return &FunctionNode{
		Name:   nodeName,
		Driver: driver,
		Args:   make(map[string]string),
	}, nil
}

func (rq *RunQueue) processLoad(bs *BlockStore) error {
	return bs.Foreach(func(b *Block) error {
		if b.Kind.Value != "load" {
			return nil
		}
		s := b.Target.Value
		fields := strings.Split(s, ":")
		dname, p, fname := fields[0], fields[1], path.Base(fields[1])
		if _, ok := rq.Locations[fname]; ok {
			return errors.New("repeat to load function: " + fname)
		}
		rq.Locations[fname] = LoadedLocation{
			DriverName:   dname,
			Path:         p,
			FunctionName: fname,
		}
		return nil
	})
}

func (rq *RunQueue) processFn(bs *BlockStore) error {
	return bs.Foreach(func(b *Block) error {
		if b.Kind.Value != "fn" {
			return nil
		}
		nodeName, fName := b.Target.Value, b.TypeOrValue.Value
		node, err := rq.createFunctionNode(nodeName, fName)
		if err != nil {
			return err
		}
		if _, ok := rq.ConfiguredNodes[node.Name]; ok {
			return errors.New("repeat to configure function:" + node.Name)
		}
		rq.ConfiguredNodes[node.Name] = node
		for _, child := range b.Child {
			if child.Kind.Value == "args" {
				node.Args = child.BlockBody.(*FlMap).ToMap()
			}
		}
		return nil
	})
}

func (rq *RunQueue) processRun(bs *BlockStore) error {
	return bs.Foreach(func(b *Block) error {
		if b.Kind.Value != "run" {
			return nil
		}
		if name := b.Target.Value; name != "" {
			// 这里属于串行调用函数 run
			//
			node, ok := rq.ConfiguredNodes[name]
			if !ok {
				// 没有显示配置函数，采用默认函数名直接run
				var err error
				if node, err = rq.createFunctionNode(name, name); err != nil {
					return err
				}
				if b.BlockBody != nil {
					node.Args = b.BlockBody.(*FlMap).ToMap()
				}
			} else {
				// 已经配置过的函数
				// TODO: 是否要处理一下 args ?
			}
			rq.Queue = append(rq.Queue, node)
			return nil
		}

		// 这里属于并行调用函数 run
		//
		var last *FunctionNode
		names := b.BlockBody.(*FlList).ToSlice()
		for _, name := range names {
			node, ok := rq.ConfiguredNodes[name]
			if !ok {
				// 没有显示配置函数，采用默认函数名直接run
				var err error
				if node, err = rq.createFunctionNode(name, name); err != nil {
					return err
				}
			}
			if last == nil {
				rq.Queue = append(rq.Queue, node)
			} else {
				last.Parallel = node
			}
			last = node
		}
		return nil
	})
}

func (rq *RunQueue) Stage(batch func(int, *FunctionNode)) {
	for i, e := range rq.Queue {
		batch(i+1, e)
	}
}
