# Performance Guidelines

パフォーマンスに関するガイドライン。

## Go パフォーマンス

### メモリ割り当て

```go
// 悪い例: ループ内で毎回割り当て
for _, item := range items {
    result = append(result, process(item))
}

// 良い例: 事前に容量確保
result := make([]T, 0, len(items))
for _, item := range items {
    result = append(result, process(item))
}
```

### 文字列連結

```go
// 悪い例: + で連結
result := ""
for _, s := range strings {
    result += s
}

// 良い例: strings.Builder
var b strings.Builder
for _, s := range strings {
    b.WriteString(s)
}
result := b.String()
```

### マップ初期化

```go
// 悪い例
m := make(map[string]int)

// 良い例: サイズが分かる場合
m := make(map[string]int, expectedSize)
```

## bon3ai 固有

### ファイルツリー

- 遅延ロード（必要時に子ノードを読み込み）
- 大きなディレクトリはページネーション検討

### VCS 操作

- VCS コマンドは非同期で実行
- 結果をキャッシュ（更新間隔を設定）

### 描画

- 変更がない場合は再描画をスキップ
- 見えない部分は描画しない

### ファイル監視

```go
// デバウンス: 短時間の連続イベントをまとめる
// watcher.go で 200ms デバウンス実装済み
```

## プロファイリング

### CPU プロファイル

```bash
go test -cpuprofile=cpu.out -bench .
go tool pprof cpu.out
```

### メモリプロファイル

```bash
go test -memprofile=mem.out -bench .
go tool pprof mem.out
```

### ベンチマーク

```go
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Function()
    }
}
```

```bash
go test -bench=. -benchmem
```

## 最適化の原則

1. **計測してから最適化**: 推測で最適化しない
2. **読みやすさ優先**: 明確な問題がなければ可読性を優先
3. **ホットパスに集中**: 頻繁に実行されるコードを最適化
4. **アルゴリズムが先**: マイクロ最適化より計算量の改善
