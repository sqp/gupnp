TARGET=guigtk
PKG=$(TARGET)

%: bindata

bindata:
	rm -f $(TARGET).xml.go
	go-bindata -pkg=$(PKG) $(TARGET).xml 
