package mdconv

type Service struct {
	detectors  []detector
	converters map[string]converter
}

func NewService() *Service {
	service := &Service{
		converters: map[string]converter{},
	}

	service.detectors = []detector{
		detectByExtension,
		detectOOXMLFormat,
		detectByContent,
	}

	service.register("txt", convertPlainText)
	service.register("md", convertPlainText)
	service.register("html", convertHTML)
	service.register("csv", convertCSV)
	service.register("json", convertJSON)
	service.register("yaml", convertYAML)
	service.register("xml", convertXML)
	service.register("zip", service.convertZIP)
	service.register("docx", convertDOCX)
	service.register("xlsx", convertXLSX)
	service.register("pptx", convertPPTX)

	return service
}

func (service *Service) Convert(data []byte, opts Options) (Result, error) {
	normalized := opts.normalized()
	format := normalized.Format
	if format == "" {
		format = service.detectFormat(normalized.Name, data)
	}

	handler, ok := service.converters[format]
	if !ok {
		return Result{}, unsupportedFormatError(format)
	}

	result, err := handler(data, normalized)
	if err != nil {
		return Result{}, err
	}

	if result.Source == "" {
		result.Source = normalized.Name
	}
	if result.Format == "" {
		result.Format = format
	}

	return result, nil
}

func (service *Service) detectFormat(name string, data []byte) string {
	for _, detect := range service.detectors {
		format, ok := detect(name, data)
		if ok {
			return format
		}
	}

	return ""
}

func (service *Service) register(format string, handler converter) {
	service.converters[format] = handler
}
