# freyjadb
A DynamoDB inspired database written with love in golang.

## Concurrency Error 
```
==================
WARNING: DATA RACE
Write at 0x00c0000a0450 by goroutine 14:
  github.com/ssargent/freyjadb/pkg/bptree.(*BPlusTree[go.shape.int,go.shape.string]).splitLeaf()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree.go:221 +0x770
  github.com/ssargent/freyjadb/pkg/bptree.(*BPlusTree[go.shape.int,go.shape.string]).Insert()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree.go:171 +0x438
  github.com/ssargent/freyjadb/pkg/bptree_test.TestBPlusTree_WriteConcurrency.func1()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree_test.go:115 +0xa4
  github.com/ssargent/freyjadb/pkg/bptree_test.TestBPlusTree_WriteConcurrency.gowrap1()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree_test.go:116 +0x44

Previous read at 0x00c0000a0450 by goroutine 10:
  github.com/ssargent/freyjadb/pkg/bptree.(*BPlusTree[go.shape.int,go.shape.string]).Insert()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree.go:133 +0x54
  github.com/ssargent/freyjadb/pkg/bptree_test.TestBPlusTree_WriteConcurrency.func1()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree_test.go:115 +0xa4
  github.com/ssargent/freyjadb/pkg/bptree_test.TestBPlusTree_WriteConcurrency.gowrap1()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree_test.go:116 +0x44

Goroutine 14 (running) created at:
  github.com/ssargent/freyjadb/pkg/bptree_test.TestBPlusTree_WriteConcurrency()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree_test.go:113 +0x28c
  testing.tRunner()
      /opt/homebrew/Cellar/go/1.22.1/libexec/src/testing/testing.go:1689 +0x180
  testing.(*T).Run.gowrap1()
      /opt/homebrew/Cellar/go/1.22.1/libexec/src/testing/testing.go:1742 +0x40

Goroutine 10 (running) created at:
  github.com/ssargent/freyjadb/pkg/bptree_test.TestBPlusTree_WriteConcurrency()
      /Users/scott/source/github/ssargent/freyjadb/pkg/bptree/bptree_test.go:113 +0x28c
  testing.tRunner()
      /opt/homebrew/Cellar/go/1.22.1/libexec/src/testing/testing.go:1689 +0x180
  testing.(*T).Run.gowrap1()
      /opt/homebrew/Cellar/go/1.22.1/libexec/src/testing/testing.go:1742 +0x40
==================
```