<h1 align="center">
  <img alt="cgapp logo" src="https://s3.bmp.ovh/imgs/2022/04/29/1ce744a5491bb3a4.png" width="200px"/><br/>
</h1>
<p align="center">KPlayer帮助你不依赖GUI快速的在服务器上进行视频资源的直播推流</p>

<p align="center"><a href="https://pkg.go.dev/github.com/create-go-app/cli/v3?tab=doc" 
target="_blank"><img src="https://img.shields.io/badge/Go-1.17+-00ADD8?style=for-the-badge&logo=go" alt="go version" /></a>&nbsp;<a href="https://gocover.io/github.com/create-go-app/cli/pkg/cgapp" target="_blank"><img src="https://img.shields.io/badge/Go_Cover-88.3%25-success?style=for-the-badge&logo=none" alt="go cover" /></a>&nbsp;<a href="https://goreportcard.com/report/github.com/create-go-app/cli" target="_blank"><img src="https://img.shields.io/badge/Go_report-A+-success?style=for-the-badge&logo=none" alt="go report" /></a>&nbsp;<img src="https://img.shields.io/badge/license-apache_2.0-red?style=for-the-badge&logo=none" alt="license" /></p>

### ⚡️ 快速开始
KPlayer可以帮助你快速的在服务器上进行视频资源的循环直播推流，只需要简单对配置文件进行自定义即可开启直播推流

> 本仓库为`libkplayer`的golang封装版本

运行以下命令即可把最新版本的kplayer下载至你的服务器上，并修改对应的`config.json`配置文件运行即可推流
```shell
# 配置文件详细文档请访问：[https://kplayer.net/p/1](https://kplayer.net/p/1) 

curl -fsSL get.kplayer.net | bash
```



###  💬 kplayer是什么
kplayer为你提供最小化成本搭建视频推流功能的工具，最优的推流方案OBS或其他软件依赖与xWindow或图形化界面的需要，不适合在服务端与云服务器上进行部署。KPlayer无需依赖图形化界面，您可以使用任意一款你喜欢的发行版本即可实现多视频资源无缝推流的方案。

只需要定义您的配置文件，针对定制化的修改。即可达成想要的结果。并且可以24小时无人值守的方式运行它。

### 🚀 有什么特性
#### 1. 多视频资源无缝推流
媒体资源推流，通常是单个资源文件连续推流。如果你有这方面的经验，那么肯定使用过`ffmpeg`或者`obs`的方案来进行推流直播。
* 与ffmpeg相比。想要实现多资源连续推流的方式通过`concat`配合`-loop`可以达到或者使用循环运行ffmpeg命令来长时间推流。但是无法动态控制视频资源的顺序，而且在视频存在差异性的情况下，必须保证视频参数的高度一致性。类似分辨率，码率，`sar`，`dar`，声道数量等造成极大的不便。使用命令行循环推流则会导致资源切换时会出现资源断流的情况，严重时会出现编码数据不匹配（绿屏、音画不同步...）
* 与obs相比。obs更依赖图形化界面的GUI操作，依赖实时编码。这在服务器上将变得不太友好。

#### 2. 预生成缓存，节省硬件资源
如果你的场景是循环推流，并不需要进行实时编码。KPlayer提供缓存机制，将上一次推流的数据缓存下来。下一次直接使用缓存文件，这将极大的降低你的机器CPU与内存占用量，仅使用较小的资源可以完成不间断推流。
KPlayer也支持在高性能机器上预生成缓存，传输至性能较小的服务器上直接使用缓存推流。降低资源占用量

#### 3. 支持多输出资源
KPlayer不仅支持输入资源的定义，对输出资源也允许定义多个输出资源并行推流。这意味着，你可以在不同的推流平台上显示一致的视频画面。
同时支持你配置重连机制，在某些原因下由于服务端的意外断开。你可以允许KPlyaer不被中断，并在某个时间段后进行尝试重新连接。

#### 4. 提供丰富的API接口辅助第三方应用控制
若具备基础的编码能力，KPlayer支持你通过`jsonrpc`调用的方式去控制它的播放行为。包括但不限于添加/删除输入资源文件、添加/删除输出资源、暂停、跳过等等等...
API是动态控制的，不必重新运行它。

#### 5. 提供可热拔插的插件机制，并提供自定义插件开发
丰富视频资源内容，我们提供了插件的机制。通过插件的配置，你可以实现插件提供的各种功能。例如在直播资源中添加一行文字、添加一个图片水印、显示时间、进度条等等...
并且支持你开发自定义插件提供给其他人使用。

###  🎨 什么样的设计
#### 1. 编码语言
KPlayer从v0.5.0以上由以下编程语言构成。
* C++17 实现编解码与输入输出的核心逻辑
* Golang 实现用户交互态的业务逻辑
* Rust/C++ 提供插件的实现功能
  相较于v0.4.x的版本，我们将各个功能解耦方便迭代开发提供更好的迭代周期和功能开发

#### 2. 解耦设计
KPlayer的整体控制逻辑依赖于消息队列通信，在主程序编码中可以看到大量的消息事件的处理，方便各模块中的功能解耦。同时多线程间彼此通过消息通信进行逻辑解耦
在对`libkplayer`与外界交互信息上，使用`protobuf`进行数据交互。如果有幸你参与到插件的开发工作中来，相信这会对你带来较大的便利

#### 3. 插件机制
得益于插件的设计逻辑，可以丰富推流视频中的内容。v0.4.x内的插件依赖于动态链接库的加载，不好的地方就是插件行为将变得不可控（读取机器磁盘文件、访问网络资源）...
得益于`WebAssembly`这项技术的优秀设计，在使用wasm来完成插件的编码与运行时，我们可以严格控制每个插件的可访问行为。在无授权的情况下，它并不能访问任何关于你机器上的任何数据。你可以放心的使用它而不必担心会存在恶意插件或插件被篡改的情况产生。并且你可以使用你熟悉的任何语言来编写插件，只要它符合wasm标准

访问这里，查看已被我们收录的插件：[https://kplayer.net/t/plugin](https://kplayer.net/t/plugin)

###  🎉 未来会支持什么
1. 提供更多的插件
2. 提供更多的辅助工具降低入门成本
3. 完善的辅助文档
