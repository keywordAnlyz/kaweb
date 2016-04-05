# kaweb
关键字分析统计系统

# 安装
命令获取相关包
```bash
go get github.com/keywordAnlyz/kaweb
```
也可以一次性更新相关包：
```bash
go get -u github.com/keywordAnlyz/kaweb
```

# 使用
1. 使用 go build 编译 kaweb .
1. 在 window 电脑下右键 【管理员权限】运行 kaweb.exe.
2. 启动成功后会显示
```bash
http server Runing on :8080
```
  
3. 打开项目主页
http://127.0.0.1:8080/ 


# 系统功能

1. 批量支持对txt,doc,docx 进行中文分词分析
2. 报表统计关键字
3. 分析人物可重复执行
4. 支持自定义字典
5. 支持自定义频次  

# 版本日志：

## v2.3 
1.支持自定义日志
2.将项目 [sego](https://github.com/huichen/sego) fork 到项目组keywordAnlyz下以便修改

## v2.2 
1. 修正IE下获取文件名不合理问题

## v2.1 
  1.修正IE下文件上传问题

## v1.5 
  1.支持tar,zip,txt,doc,docx 文件上传