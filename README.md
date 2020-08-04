# mcmober
A tool for download mobs from my mc server

运行平台是windows.
之前租了个机器器搭建了自己的mc服务器, 使用这个小工具从服务器上下载mob.
流程很简单, 就是获取文件列表, 比对, 然后下载不同或者不存在的mob而已.

syso.bat, 使用windres, 通过rc文件生成syso文件, 给程序带上图标
mcmober.go, 程序主文件
mcmober.ini, 配置文件, 主要配置mc服务器地址和端口
