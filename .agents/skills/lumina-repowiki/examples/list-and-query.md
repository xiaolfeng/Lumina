# Example: 列出完成版本并读架构页

Input: 刚进陌生仓库，想先看模块划分再改代码。

## 1. 列表

```json
{ "page": 1, "size": 20 }
```

```text
Wiki 版本列表（共 1 个，第 1/1 页）：

1. [version_id: 1234567890123456789] Lumina
   分支: master | 语言: Go | commit: 596abd9
   完成时间: 2026-08-26T10:00:00Z
```

记下 `version_id`，当作后续 `wiki_id`。

## 2. 首页

```json
{ "wiki_id": 1234567890123456789 }
```

从首页目录找到架构章节路径。

## 3. 指定页面

```json
{
  "wiki_id": 1234567890123456789,
  "page": "content/architecture"
}
```

## 反例

```json
{ "wiki_id": 1234567890123456789, "page": "content/architecture.mdx" }
```

带扩展名会找不到页。

不要调用不存在的 `repoWiki_analyze` / `repoWiki_update`。生成由 Git Push + Webhook 驱动。
