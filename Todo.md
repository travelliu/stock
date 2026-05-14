# ToDO

- [analysis](pkg/analysis) 
  - 只保留分析的代码就行了. 不需要存打印的. 打印是ctl的 放到ctl的代码包里
  - 后端代码里的中文改成key 如		"", "历史参考价", "近3月参考价", "近1月参考价", "近2周参考价", 供生成工具编译前生成多语言映射给前端和后端用. 并统一考虑下多语言.
- pkg/shared 这里里面不需要在拆分模块了. 有些数据可以放到analysis里了
  - ToTushareCode 都可以当做工具包不需要额外的
  - [spread.go](pkg/shared/spread/spread.go) 
    - 和tushare的数据放到tushare
    - 自己的数据放大models里
    - models提取到 pkg/下,方便 stockd/ctl使用
