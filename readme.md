# `golang`使用`gdal`库

为`golang`编译 `gdal` `geos`库文件

## 编译`gdal`库
- 不太懂C++，我使用的是`vcpkg`编译，具体[看这里](https://gdal.org/download.html#vcpkg)
- `vcpkg`的介绍自行百度学习或者[看这里](https://github.com/microsoft/vcpkg/blob/master/README_zh_CN.md#%E5%9C%A8-cmake-%E4%B8%AD%E4%BD%BF%E7%94%A8-vcpkg)
- 编译时间有点长（一个小时左右），不想自己编译的直接下载我这个仓库就好了。编译的gdal版本为3.7.1

```bat
rem 克隆vcpkg库
git clone git@github.com:microsoft/vcpkg.git

cd vcpkg
rem 下载vcpkg.exe程序
bootstrap-vcpkg.bat

rem 下载完成之后开始编译
vcpkg search gdal

rem 默认编译32位 加 :x64-windows 编译64位的
vcpkg install gdal
vcpkg install gdal:x64-windows
```

## 测试C语言环境
- 我是把编译好的头文件放到`mingw`对应目录下，但是编译c语言测试程序时能找到头文件，找不到库文件，加了编译参数也不行，但是在`vs`里面就没问题。

- 我的编译命令是这个``` g++ -I D:\DEV\gdal\include -LD:\DEV\gdal\lib -lgdal  -o mainc.exe main.c ``` ，会报这个错：`undefined reference to 'GDALAllRegister'` 或者这个``` ld: cannot find -lgdal ```,未解决。

- 如果测试go环境时要把main.c注释，不然go编译器会一起编译c语言

测试C/C++代码:
```C++
// main.cpp
#include <sti.h>
#include <stdio.h>
#include <gdal.h>
#include <iostream>
#include <gdal_priv.h>
#include <ogrsf_frmts.h>
#include <gdal_alg.h>


int main()
{
	const char* image_name = "E:/testtif/testtif.tif";
	GDALAllRegister();
	GDALDataset* poSrc = (GDALDataset*)GDALOpen(image_name, GA_ReadOnly);
	if (poSrc == nullptr) {
		std::cout << "input image error" << std::endl;
		return -1;
	}

	int width_src = poSrc->GetRasterXSize();
	int height_src = poSrc->GetRasterYSize();
	int band_count_src = poSrc->GetRasterCount();
	printf(stderr, "width: %d, height: %d, bandCount: %d\n", width_src, height_src, band_count_src);
	GDALDataType  gdal_data_type = poSrc->GetRasterBand(1)->GetRasterDataType();
	int depth = GDALGetDataTypeSize((GDALDataType)gdal_data_type);
	printf(stderr, "depth: %d\n", depth);

	GDALClose((GDALDatasetH)poSrc);

	return 0;
}
```
```C
// main.c
#include "gdal.h"
#include "cpl_conv.h" /* for CPLMalloc() */

#include <errno.h>

int main(int argc, const char* argv[])
{
    if (argc != 2) {
        return EINVAL;
    }
    const char* pszFilename = argv[1];

    GDALDatasetH  hDataset;
    GDALAllRegister();
    const GDALAccess eAccess = GA_ReadOnly;
    hDataset = GDALOpen( pszFilename, eAccess );
    if( hDataset == NULL )
    {
        return -1; // handle error
    }
    return 0;
}
```


## 配置`golang` `gdal`开发设置
- 下载`gdal`的go语言绑定库 ``` go get github.com/lukeroth/gdal ```
- 修改这个库的``` c_windows_amd64.go``` 文件
- 这个文件在```gopath```路径下的```pkg\mod\github.com\lukeroth\gdal@···```,一般是在``` C:\Users\username\go\pkg\mod\github.com\lukeroth\gdal@··· ```

```go
//go:build windows && amd64
// +build windows,amd64

package gdal

/*
#cgo windows CFLAGS: -ID:/DEV/gdal/include
#cgo windows LDFLAGS: -LD:/DEV/gdal/lib -lgdal -ltiff -lgeotiff_i -lgeos_c
*/
import "C"
```
- 把 CFLAGS 改成对应编译好的`gdal` 库的`include`文件夹
- 把 LDFLAGS 改成对应编译好的`gdal` 库的 `lib` 文件夹，必须加上`-lgdal -lgeos_c`参数，不然编译也是会报这个错：`undefined reference to 'GDALAllRegister'`。
- 下面是编译出来的lib文件，要是再报找不到某个函数就尝试加 `-l` 参数应该能解决
```txt
bz2
charset
freexl
gdal
geos
geos_c
geotiff_i
gif
hdf5
hdf5_cpp
hdf5_hl
hdf5_hl_cpp
iconv
jpeg
json-c-static
json-c
kmlbase
kmlconvenience
kmldom
kmlengine
kmlregionator
kmlxsd
Lerc
libcrypto
libcurl
libecpg
libecpg_compat
libexpat
libhdf5.settings
libnetcdf.settings
libpgcommon
libpgport
libpgtypes
libpng16
libpq
libsharpyuv
libssl
libwebp
libwebpdecoder
libwebpdemux
libwebpmux
libxml2
lz4
lzma
minizip
netcdf
openjp2
pcre2-16
pcre2-32
pcre2-8
pcre2-posix
pkgconf
proj
qhullcpp
qhull_r
spatialite
sqlite3
szip
tiff
turbojpeg
uriparser
zlib
zstd
```



go环境测试代码
```go
package main

import (
	"fmt"
	"flag"
	gdal "github.com/lukeroth/gdal "
)

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	if filename == "" {
		fmt.Printf("Usage: tiff [filename]\n")
		return
	}
	buffer := make([]uint8, 256 * 256)

	driver, err := gdal.GetDriverByName("GTiff")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dataset := driver.Create(filename, 256, 256, 1, gdal.Byte, nil)
	defer dataset.Close()

	spatialRef := gdal.CreateSpatialReference("")
	spatialRef.FromEPSG(3857)
	srString, err := spatialRef.ToWKT()
	dataset.SetProjection(srString)
	dataset.SetGeoTransform([]float64{444720, 30, 0, 3751320, 0, -30})
	raster := dataset.RasterBand(1)
	raster.IO(gdal.Write, 0, 0, 256, 256, buffer, 256, 256, 0, 0)
}
```

## 编译`geos`库
- 编译的gdal应该是包含了geos库了，具体没测试过，先这样把。
- 


