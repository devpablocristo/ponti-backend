package pkgexcel

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type Service struct {
	client  *excelize.File
	address string
	config  ConfigPort
}

// newService inicializa un repositorio y conecta (igual a pkggorm)
func newService(c ConfigPort) (*Service, error) {
	repo := &Service{config: c}
	if err := repo.Connect(c); err != nil {
		return nil, fmt.Errorf("failed to initialize Service: %w", err)
	}
	return repo, nil
}

// Connect abre o crea el archivo y garantiza la hoja
func (r *Service) Connect(config ConfigPort) error {
	if err := r.createWorkbookIfNotExists(config); err != nil {
		return fmt.Errorf("failed to create workbook: %w", err)
	}
	f, err := excelize.OpenFile(config.GetFilePath())
	if err != nil {
		return fmt.Errorf("open workbook: %w", err)
	}
	r.client = f
	r.address = config.GetFilePath()
	// asegurar hoja
	idx, err := r.client.GetSheetIndex(config.GetSheet())
	if err != nil || idx < 0 {
		idx, err = r.client.NewSheet(config.GetSheet())
		if err != nil {
			return fmt.Errorf("create sheet %q: %w", config.GetSheet(), err)
		}
	}
	r.client.SetActiveSheet(idx)
	return nil
}

// Client retorna el puntero excelize.File
func (r *Service) Client() *excelize.File { return r.client }

// Address retorna la ruta del archivo (igual a pkggorm.Address)
func (r *Service) Address() string { return r.address }

// Close guarda y cierra el archivo
func (r *Service) Close() error {
	if r.client == nil {
		return nil
	}
	if err := r.client.Save(); err != nil {
		return err
	}
	return r.client.Close()
}

// Export escribe `data` a la hoja configurada
func (r *Service) Export(data any) error {
	if r.client == nil {
		return errors.New("client not connected")
	}
	return exportToFile(r.client, r.config, data)
}

// ExportToWriter escribe a un writer (útil HTTP/S3). Crea un file temporal in-memory y vuelca.
func (r *Service) ExportToWriter(data any, w io.Writer) error {
	f := excelize.NewFile()
	defer f.Close()
	if err := exportToFile(f, r.config, data); err != nil {
		return err
	}
	_, err := f.WriteTo(w)
	return err
}

// Import lee la hoja y mapea a dest (puntero a slice de struct o map)
func (r *Service) Import(dest any) error {
	if r.client == nil {
		return errors.New("client not connected")
	}
	rows, err := r.client.GetRows(r.config.GetSheet())
	if err != nil {
		return fmt.Errorf("read rows: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}
	return mapRowsToDest(r.config, rows, dest)
}

// createWorkbookIfNotExists crea el archivo si no existe
func (r *Service) createWorkbookIfNotExists(config ConfigPort) error {
	if _, err := os.Stat(config.GetFilePath()); os.IsNotExist(err) {
		f := excelize.NewFile()
		defer f.Close()
		idx, err := f.NewSheet(config.GetSheet())
		if err != nil {
			return fmt.Errorf("new sheet %q: %w", config.GetSheet(), err)
		}
		f.SetActiveSheet(idx)
		if err := f.SaveAs(config.GetFilePath()); err != nil {
			return fmt.Errorf("save new workbook: %w", err)
		}
	}
	return nil
}

func exportToFile(f *excelize.File, cfg ConfigPort, data any) error {
	// crear/seleccionar hoja
	idx, err := f.GetSheetIndex(cfg.GetSheet())
	if err != nil || idx < 0 {
		idx, err = f.NewSheet(cfg.GetSheet())
		if err != nil {
			return fmt.Errorf("new sheet %q: %w", cfg.GetSheet(), err)
		}
	}
	f.SetActiveSheet(idx)

	headers, rows, err := buildRows(cfg, data)
	if err != nil {
		return err
	}

	rowOffset := 1
	if cfg.GetWriteHeader() {
		for c, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(c+1, rowOffset)
			_ = f.SetCellValue(cfg.GetSheet(), cell, h)
		}
		rowOffset++
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+rowOffset)
			_ = f.SetCellValue(cfg.GetSheet(), cell, v)
		}
	}
	if widths := cfg.GetColumnWidths(); len(widths) > 0 {
		for i, w := range widths {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = f.SetColWidth(cfg.GetSheet(), col, col, w)
		}
	}
	return nil
}

func buildRows(cfg ConfigPort, data any) ([]string, [][]any, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, nil, fmt.Errorf("data must be slice or array")
	}
	if v.Len() == 0 {
		return nil, nil, fmt.Errorf("data slice is empty")
	}
	el := v.Index(0)
	if el.Kind() == reflect.Ptr {
		el = el.Elem()
	}
	switch el.Kind() {
	case reflect.Struct:
		return rowsFromStructs(cfg, v)
	case reflect.Map:
		return rowsFromMaps(v)
	default:
		return nil, nil, fmt.Errorf("unsupported element kind: %s", el.Kind())
	}
}

func rowsFromStructs(cfg ConfigPort, v reflect.Value) ([]string, [][]any, error) {
	t := v.Index(0).Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	nf := t.NumField()
	idx := make([]int, 0, nf)
	headers := make([]string, 0, nf)
	for i := 0; i < nf; i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if tag := f.Tag.Get("excel"); tag == "-" {
			continue
		}
		name := f.Tag.Get("excel")
		if name == "" {
			name = f.Tag.Get("json")
		}
		if name == "" || name == "-" {
			name = f.Name
		}
		headers = append(headers, name)
		idx = append(idx, i)
	}
	rows := make([][]any, v.Len())
	for r := 0; r < v.Len(); r++ {
		row := make([]any, len(idx))
		item := v.Index(r)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		for c, fi := range idx {
			row[c] = toCellValue(cfg, item.Field(fi))
		}
		rows[r] = row
	}
	return headers, rows, nil
}

func rowsFromMaps(v reflect.Value) ([]string, [][]any, error) {
	keySet := map[string]struct{}{}
	for i := 0; i < v.Len(); i++ {
		it := v.Index(i)
		if it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		for _, k := range it.MapKeys() {
			keySet[k.String()] = struct{}{}
		}
	}
	headers := make([]string, 0, len(keySet))
	for k := range keySet {
		headers = append(headers, k)
	}
	sort.Strings(headers)
	rows := make([][]any, v.Len())
	for r := 0; r < v.Len(); r++ {
		row := make([]any, len(headers))
		it := v.Index(r)
		if it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		for c, hk := range headers {
			val := it.MapIndex(reflect.ValueOf(hk))
			row[c] = toCellValue(nil, val)
		}
		rows[r] = row
	}
	return headers, rows, nil
}

func toCellValue(cfg ConfigPort, v reflect.Value) any {
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Bool:
		return v.Bool()
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			dateFmt := "2006-01-02"
			if cfg != nil && cfg.GetDateFormat() != "" {
				dateFmt = cfg.GetDateFormat()
			}
			return t.Format(dateFmt)
		}
	}
	return fmt.Sprintf("%v", v.Interface())
}

func mapRowsToDest(cfg ConfigPort, rows [][]string, dest any) error {
	if len(rows) == 0 {
		return errors.New("no rows")
	}
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("dest must be non-nil pointer to slice")
	}
	slice := v.Elem()
	if slice.Kind() != reflect.Slice {
		return errors.New("dest must be pointer to slice")
	}
	headers := rows[0]
	dataRows := rows[1:]
	elType := slice.Type().Elem()
	isPtr := false
	if elType.Kind() == reflect.Ptr {
		isPtr = true
		elType = elType.Elem()
	}
	switch elType.Kind() {
	case reflect.Struct:
		fieldMap := buildFieldMap(elType)
		for _, r := range dataRows {
			el := reflect.New(elType).Elem()
			for i, h := range headers {
				if fi, ok := fieldMap[strings.ToLower(h)]; ok {
					if i < len(r) {
						if err := setFieldFromString(cfg, el.Field(fi), r[i]); err != nil {
							return err
						}
					}
				}
			}
			if isPtr {
				slice = reflect.Append(slice, el.Addr())
			} else {
				slice = reflect.Append(slice, el)
			}
		}
	case reflect.Map:
		if elType.Key().Kind() != reflect.String {
			return errors.New("map key must be string")
		}
		for _, r := range dataRows {
			m := reflect.MakeMapWithSize(elType, len(headers))
			for i, h := range headers {
				if i < len(r) {
					m.SetMapIndex(reflect.ValueOf(h), reflect.ValueOf(r[i]))
				}
			}
			slice = reflect.Append(slice, m)
		}
	default:
		return fmt.Errorf("unsupported slice element kind: %s", elType.Kind())
	}
	v.Elem().Set(slice)
	return nil
}

func buildFieldMap(t reflect.Type) map[string]int {
	m := make(map[string]int)
	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if tag := f.Tag.Get("excel"); tag == "-" {
			continue
		}
		name := f.Tag.Get("excel")
		if name == "" {
			name = f.Tag.Get("json")
		}
		if name == "" || name == "-" {
			name = f.Name
		}
		m[strings.ToLower(name)] = i
	}
	return m
}

func setFieldFromString(cfg ConfigPort, f reflect.Value, val string) error {
	if !f.CanSet() {
		return nil
	}
	if f.Kind() == reflect.Ptr {
		if val == "" {
			return nil
		}
		ptr := reflect.New(f.Type().Elem())
		if err := setFieldFromString(cfg, ptr.Elem(), val); err != nil {
			return err
		}
		f.Set(ptr)
		return nil
	}
	switch f.Kind() {
	case reflect.String:
		f.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		f.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		f.SetUint(u)
	case reflect.Float32, reflect.Float64:
		fl, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		f.SetFloat(fl)
	case reflect.Bool:
		b, err := strconv.ParseBool(strings.ToLower(val))
		if err != nil {
			return err
		}
		f.SetBool(b)
	case reflect.Struct:
		if f.Type() == reflect.TypeOf(time.Time{}) {
			df := "2006-01-02"
			if cfg != nil && cfg.GetDateFormat() != "" {
				df = cfg.GetDateFormat()
			}
			t, err := time.Parse(df, val)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(t))
			return nil
		}
		return fmt.Errorf("unsupported struct type: %s", f.Type())
	default:
		return fmt.Errorf("unsupported kind: %s", f.Kind())
	}
	return nil
}
