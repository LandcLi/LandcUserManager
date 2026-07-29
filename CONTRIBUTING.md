# Contributing to LandcUserManager

感谢您考虑为 LandcUserManager 做出贡献！

## 开发流程

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feat/your-feature`
3. 提交改动：`git commit -m 'feat: add some feature'`
4. 推送分支：`git push origin feat/your-feature`
5. 创建 Pull Request

## 代码规范

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 所有导出函数必须有注释
- 错误处理：使用 `pkg/errors` 中的预定义错误变量，而非硬编码字符串
- 魔法字符串：使用 `pkg/consts` 中的常量
- 提交前运行 `make lint && make test`

## 提交信息规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

- `feat`: 新功能
- `fix`: 修复
- `docs`: 文档
- `refactor`: 重构
- `test`: 测试
- `chore`: 杂项

## 测试

- 新增功能需包含单元测试
- 运行所有测试：`make test`
