# Contributing to LinuxDo-Relay

感谢你对 LinuxDo-Relay 的关注！我们欢迎各种形式的贡献。

## 如何贡献

### 报告 Bug

如果你发现了 bug，请创建一个 Issue，包含以下信息：

- **清晰的标题**：简短描述问题
- **详细描述**：问题的详细描述
- **复现步骤**：如何复现这个问题
- **期望行为**：你期望发生什么
- **实际行为**：实际发生了什么
- **环境信息**：
  - 操作系统和版本
  - Go 版本
  - 数据库版本
  - 浏览器版本（如果是前端问题）
- **截图/日志**：如果有的话

### 提出新功能

如果你有新功能的想法，请创建一个 Issue，包含：

- **功能描述**：清晰描述新功能
- **使用场景**：为什么需要这个功能
- **可能的实现方式**：如果有想法的话

### 提交代码

1. **Fork 项目**
   ```bash
   # 点击 GitHub 上的 Fork 按钮
   git clone https://github.com/YOUR_USERNAME/linuxdo-relay.git
   cd linuxdo-relay
   ```

2. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   # 或
   git checkout -b fix/your-bug-fix
   ```

3. **开发和测试**
   ```bash
   # 后端测试
   go test ./...
   
   # 前端测试
   cd web && npm run test
   ```

4. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   # 或
   git commit -m "fix: resolve issue with X"
   ```

   遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：
   - `feat`: 新功能
   - `fix`: Bug 修复
   - `docs`: 文档更新
   - `style`: 代码格式化
   - `refactor`: 代码重构
   - `test`: 测试相关
   - `chore`: 构建/工具相关

5. **推送到你的 Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **创建 Pull Request**
   - 在 GitHub 上打开你的 Fork
   - 点击 "New Pull Request"
   - 填写 PR 描述，说明你的更改

## 开发规范

### 后端 (Go)

- 遵循 Go 官方代码风格
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 检查代码质量
- 编写单元测试
- 添加必要的注释

### 前端 (React)

- 遵循 Airbnb JavaScript 风格指南
- 使用 ESLint 和 Prettier
- 组件使用函数式组件和 Hooks
- 编写组件测试

### 提交信息

良好的提交信息示例：

```
feat: add daily check-in feature

- Add check-in API endpoints
- Implement reward calculation with decay
- Add check-in UI in user dashboard
- Update database schema with migration 004

Closes #123
```

## 代码审查

所有的 PR 都需要经过代码审查。审查者会检查：

- 代码质量和可读性
- 是否有测试覆盖
- 是否符合项目架构
- 文档是否完整

## 许可证

提交代码即表示你同意将你的贡献按照项目的 LICENSE 许可。

## 获取帮助

- 📖 查看 [README.md](./README.md)
- 📋 查看 [ADMIN_GUIDE.md](./ADMIN_GUIDE.md)
- 💬 在 [Discussions](https://github.com/dbccccccc/linuxdo-relay/discussions) 提问
- 🐛 在 [Issues](https://github.com/dbccccccc/linuxdo-relay/issues) 报告问题

感谢你的贡献！🎉
