package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
)

func Load(configDir string) {
	//加载主配置文件.
	if err := loadMasterConf(configDir); err != nil {
		log.Printf("err:%+v\n", err)
		return
	}

	//加载其他配置文件.
	files, err := getConfigFile(configDir)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return
	}

	readConf(files)
}

func getConfigFile(dirPth string) (files []string, err error) {
	fis, err := ioutil.ReadDir(filepath.Clean(filepath.ToSlash(dirPth)))
	if err != nil {
		return nil, err
	}

	for _, f := range fis {
		_path := filepath.Join(dirPth, f.Name())

		if f.IsDir() {
			continue
		}

		// 指定格式
		switch filepath.Ext(f.Name()) {
		case ".yaml":
			files = append(files, _path)
		}
	}

	return files, nil
}

func readConf(files []string) {
	for _, _path := range files {
		if err := readConfItem(_path); err != nil {
			log.Printf("err:%+v\n", err)
			return
		}
	}
}

func readConfItem(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	cf := WorkServer{}
	err = yaml.Unmarshal(buf, &cf)
	if err != nil {
		return err
	}

	switch cf.Type {
	case ConfigType_FileServer:
		ConfigMgr.FileServer = append(ConfigMgr.FileServer, FileServerConfig{
			Type:       cf.Type,
			ServerName: cf.ServerName,
			Index:      cf.Index,
			Root:       cf.Root,
		})
	case ConfigType_ReverseProxy:
		ConfigMgr.ReverseProxy = append(ConfigMgr.ReverseProxy, ReverseProxyConfig{
			Type:       cf.Type,
			ServerName: cf.ServerName,
			ProxyPass:  cf.ProxyPass,
			Module:     cf.Module,
		})
	}
	return nil
}

//loadMasterConf 加载主配置文件.
func loadMasterConf(configDir string) error {
	masterConfigPath := path.Clean(fmt.Sprintf("%s/workserver.yaml", configDir))
	buf, err := ioutil.ReadFile(masterConfigPath)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return err
	}
	cf := &Config{}

	if err := yaml.Unmarshal(buf, cf); err != nil {
		log.Printf("err:%+v\n", err)
		return err
	}

	ConfigMgr = cf
	return nil
}
