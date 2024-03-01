package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

var errUnknownPattern = errors.New("未知的双写模式")

const (
	PatternDstOnly  = "DST_ONLY"
	PatternSrcOnly  = "SRC_ONLY"
	PatternDstFirst = "DST_FIRST"
	PatternSrcFirst = "SRC_FIRST"
)

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
}

func NewDoubleWritePool(src gorm.ConnPool,
	dst gorm.ConnPool, pattern string) *DoubleWritePool {
	return &DoubleWritePool{src: src, dst: dst, pattern: atomicx.NewValueOf(pattern)}
}

// PrepareContext Prepare 的语句会进来这里
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("implement me")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	default:
		panic("未知的双写模式")
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 这里有一个问题，我怎么返回一个 error
		// unsafe 可以
		panic("未知的双写模式")
	}
}

// UpdatePattern 更新双写模式
func (d *DoubleWritePool) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

type DoubleWritePoolTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
}

// Commit 和 PPT 不一致
func (d *DoubleWritePoolTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		// 源库上的事务失败了，我目标库要不要提交
		// commit 失败了怎么办？
		err := d.src.Commit()
		if err != nil {
			// 要不要提交？
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			// 要不要提交？
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		// 源库上的事务失败了，我目标库要不要提交
		// commit 失败嘞怎么办？
		err := d.src.Rollback()
		if err != nil {
			// 要不要提交？
			// 我个人觉得 可以尝试 rollback
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			// 要不要提交？
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("implement me")
}

func (d *DoubleWritePoolTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.dst == nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.src == nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePoolTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePoolTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}
