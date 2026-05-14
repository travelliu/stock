// Package render formats analysis output for terminal display.
package render

import (
	"fmt"
	"os"
	"stock/pkg/models"
	"text/tabwriter"
)

func AnalysisTable(r models.AnalysisResult) {
	fmt.Printf("\n%s (%s)\n", r.StockName, r.TsCode)
	if r.YesterdayClose != nil {
		fmt.Printf("昨收: %.2f\n\n", *r.YesterdayClose)
	} else {
		fmt.Printf("昨收: -\n\n")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, join(r.ModelTable.Headers, "\t")+"\t")
	for _, row := range r.ModelTable.Rows {
		fmt.Fprintln(w, join(row, "\t")+"\t")
	}
	w.Flush()

	fmt.Println("\n参考区间:")
	w2 := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w2, join(r.ReferenceTable.Headers, "\t")+"\t")
	for _, row := range r.ReferenceTable.Rows {
		fmt.Fprintln(w2, join(row, "\t")+"\t")
	}
	w2.Flush()
}

func join(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for _, s := range ss[1:] {
		out += sep + s
	}
	return out
}
