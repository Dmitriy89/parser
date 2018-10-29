package myreports

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type chartline struct {
	date         []string
	brand        string
	countProduct []string
}

func (c *chartline) Append(s string, i string) {
	c.date = append(c.date, s)
	c.countProduct = append(c.countProduct, i)
}

// RGBA return color chart
func RGBA(count int) []string {
	var (
		slice []string
		f     func() string
	)
	f = func() string {
		time.Sleep(time.Microsecond * 1)
		s := rand.NewSource(time.Now().UnixNano())
		random := rand.New(s).Uint32()
		return strconv.Itoa(int(uint8(random)))
	}
	for i := 0; i < count; i++ {
		slice = append(slice, "'rgba("+f()+","+f()+","+f()+","+"0.4)',")
	}
	return slice
}

//ChartHBar create chart horizontalBar
func ChartHBar(st []time.Time) error {
	var w, h int

	countBrandBD, err := DataChartHBarDB(st)
	if err != nil {
		return fmt.Errorf("Error func ChartHBar func DataChartDB:__ %s", err)
	}

	html, err := os.Create("indexHBar.html")
	if err != nil {
		return fmt.Errorf("Error func ChartHBar method Create:__ %s", err)
	}

	sliceBrandHTML := make([]string, 0)
	sliceCountBrandHTML := make([]string, 0)
	color := RGBA(len(countBrandBD) / 2)

	w = 450
	h = 450

	if len(countBrandBD)/2 > 30 {
		w = 750
		h = 750
	}
	for key, val := range countBrandBD {
		if key%2 == 0 {
			sliceBrandHTML = append(sliceBrandHTML, `"`+val+`",`)
		} else {
			sliceCountBrandHTML = append(sliceCountBrandHTML, val+`,`)
		}

	}

	s := fmt.Sprintf(`<!DOCTYPE html>
			<html>
			<head>
				<title>ЦУМ</title>
				<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
				<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.7.2/Chart.bundle.min.js"></script>
				<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-datalabels"></script>

			</head>
			<body>
				<canvas id="Chart" width="%d" height="%d"></canvas>
				<script>
					var ctx = document.getElementById("Chart").getContext('2d');
					var myChart = new Chart(ctx, {
						type: 'horizontalBar',
						data: {
							labels:%s,
							datasets: [{
								label: ['Brand statistics'],
								data: %s,
								backgroundColor: %s,
							}]
						},
						options: {
							responsive: false,
							legend: {
								display: false
							}
						}
					});
				</script>
			</body>
			</html>`, w, h, sliceBrandHTML, sliceCountBrandHTML, color)
	read := strings.NewReader(s)
	io.Copy(html, read)

	return nil
}

//ChartLineGeneral create chart line
func ChartLineGeneral(st []time.Time) error {
	var maxDate []string
	var arrayOption string
	var sliceArrayOption []string

	sliceDB, err := DataChartLineDB(st)
	if err != nil {
		return fmt.Errorf("Error func ChartLine method ReadDB:__ %s", err)
	}

	html, err := os.Create("indexLineGeneral.html")
	if err != nil {
		return fmt.Errorf("Error func ChartLine method Create:__ %s", err)
	}

	sliceHash := make(map[string]*chartline)

	for _, val := range sliceDB {

		if _, ok := sliceHash[val[1]]; !ok {
			sliceHash[val[1]] = &chartline{[]string{val[0]}, val[1], []string{val[2] + ","}}
		} else {
			sliceHash[val[1]].Append(val[0], val[2]+",")
		}
	}

	for _, val := range sliceHash {
		if len(maxDate) < len(val.date) {
			maxDate = make([]string, len(val.date))
			copy(maxDate, val.date)
		}
		color := RGBA(len(sliceHash))
		s := fmt.Sprintf(`{
			label: "%s",
			data: %s,
			fill: false,
			borderColor: %s				
		}`, val.brand, val.countProduct, color[0])
		sliceArrayOption = append(sliceArrayOption, s)
	}
	arrayOption = strings.Join(sliceArrayOption, ",")
	//maxDate = append(maxDate, "\"\"")
	s := fmt.Sprintf(`<!DOCTYPE html>
			<html>
			<head>
				<title>ЦУМ</title>
				<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
				<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.7.2/Chart.bundle.min.js"></script>
				<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-datalabels"></script>

			</head>
			<body>
				<canvas id="Chart" width="%d" height="%d"></canvas>
				<script>
					var ctx = document.getElementById("Chart").getContext('2d');
					var myChart = new Chart(ctx, {
						type: 'line',
						data: {
							labels:%s,
							datasets: [%s]
						},
						options: {
							responsive: false,
							legend: {
								display: true,
								position: "right"
							},				
							scales: {
								yAxes: [{
									ticks: {
										beginAtZero: true,
										stepSize: .5,
									}
								}]
							}
						}
					});
				</script>
			</body>
			</html>`, 1500, 550, maxDate, arrayOption)
	read := strings.NewReader(s)
	io.Copy(html, read)

	return nil
}

//ChartLineIndividual create chart line
func ChartLineIndividual(st []time.Time) error {
	var (
		maxDate     []string
		arrayScript []string
		slicecanvas []string
	)

	sliceDB, err := DataChartLineDB(st)
	if err != nil {
		return fmt.Errorf("Error func ChartLine method ReadDB:__ %s", err)
	}

	html, err := os.Create("indexLineIndividual.html")
	if err != nil {
		return fmt.Errorf("Error func ChartLine method Create:__ %s", err)
	}

	sliceHash := make(map[string]*chartline)

	for _, val := range sliceDB {

		if _, ok := sliceHash[val[1]]; !ok {
			sliceHash[val[1]] = &chartline{[]string{val[0]}, val[1], []string{val[2] + ","}}
		} else {
			sliceHash[val[1]].Append(val[0], val[2]+",")
		}
	}
	i := 0
	for _, val := range sliceHash {
		if len(maxDate) < len(val.date) {
			maxDate = make([]string, len(val.date))
			copy(maxDate, val.date)
		}
		iteration := strconv.Itoa(i)
		slicecanvas = append(slicecanvas, "<canvas id='Chart"+iteration+"' width='1000' height='170'></canvas><br />")
		s := fmt.Sprintf(`
		var ctx%s = document.getElementById("Chart%s").getContext('2d');
		var myChart%s = new Chart(ctx%s, {
			type: 'line',
			data: {
				labels: %s,
				datasets: [{
					label: "%s",
					data: %s,
					fill: false,
					borderColor: 'rgba(231,156,59,0.4)',
					backgroundColor: 'rgba(231,156,59,0.7)'
				}]				
			},
			options: {
				plugins: {
					datalabels: {
						anchor: "start",
						align: "end"
					}
				},
				responsive: true,
				legend: {
					display: true,
					position: "top"
				},
				scales: {
					yAxes: [{
						ticks: {
							beginAtZero: true,
							stepSize: 3,
							max: 20
						}
					}]
				}
			}
		})`,
			iteration, iteration, iteration, iteration, maxDate, val.brand, val.countProduct)
		arrayScript = append(arrayScript, s)
		i++
	}
	canvas := strings.Join(slicecanvas, " ")
	script := strings.Join(arrayScript, ";")
	//maxDate = append(maxDate, "\"\"")
	s := fmt.Sprintf(`<!DOCTYPE html>
			<html>
			<head>
				<title>ЦУМ</title
				<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
				<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.7.2/Chart.bundle.min.js"></script>
				<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-datalabels"></script>

			</head>
			<body>
				<center><h1>Новинки</h1></center>
				%s
				<script>
				%s
				</script>
			</body>
			</html>`, canvas, script)
	read := strings.NewReader(s)
	io.Copy(html, read)

	return nil
}
