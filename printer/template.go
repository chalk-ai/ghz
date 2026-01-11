package printer

var (
	defaultTmpl = `
Summary:
{{ if .Name }}  Name:		{{ .Name }}
{{ end }}  Count:	{{ .Count }}
  Total:	{{ formatNanoUnit .Total }}
  Slowest:	{{ formatNanoUnit .Slowest }}
  Fastest:	{{ formatNanoUnit .Fastest }}
  Average:	{{ formatNanoUnit .Average }}
  Requests/sec:	{{ formatSeconds .Rps }}

Response time histogram:
{{ histogram .Histogram }}
Latency distribution:{{ range .LatencyDistribution }}
  {{ .Percentage }} % in {{ formatNanoUnit .Latency }} {{ end }}

{{ if gt (len .StatusCodeDist) 0 }}Status code distribution:
{{ formatStatusCode .StatusCodeDist }}{{ end }}
{{ if gt (len .ErrorDist) 0 }}Error distribution:
{{ formatErrorDist .ErrorDist }}{{ end }}
`

	csvTmpl = `
duration (ms),status,error{{ range $i, $v := .Details }}
{{ formatMilli .Latency.Seconds }},{{ .Status }},{{ .Error }}{{ end }}
`

	htmlTmpl = `
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Chalk Benchmark {{ if .Name }} - {{ .Name }}{{end}}</title>
  	<script src="https://cdn.jsdelivr.net/npm/papaparse@4.5.0/papaparse.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/apache-arrow@18.1.0/Arrow.es2015.min.js"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.7.1/css/bulma.min.css" />
    <style>
      .json-key { color: #881391; }
      .json-string { color: #1a1aa6; }
      .json-number { color: #1c00cf; }
      .json-boolean { color: #0d22ff; }
      .json-null { color: #808080; }

      /* Toggle switch styles */
      .switch {
        position: relative;
        display: inline-block;
        width: 42px;
        height: 22px;
        vertical-align: middle;
      }

      .switch input {
        opacity: 0;
        width: 0;
        height: 0;
      }

      .switch-slider {
        position: absolute;
        cursor: pointer;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: #dbdbdb;
        transition: .3s;
        border-radius: 22px;
      }

      .switch-slider:before {
        position: absolute;
        content: "";
        height: 16px;
        width: 16px;
        left: 3px;
        bottom: 3px;
        background-color: white;
        transition: .3s;
        border-radius: 50%;
      }

      .switch input:checked + .switch-slider {
        background-color: #3273dc;
      }

      .switch input:checked + .switch-slider:before {
        transform: translateX(20px);
      }

      /* Table styles */
      #dataTable {
        width: 100%;
        overflow-x: auto;
        overflow-y: visible;
      }

      #dataTable table {
        width: 100%;
        border-collapse: collapse;
        background-color: white;
      }

      #dataTable thead {
        background-color: #f5f5f5;
      }

      #dataTable th {
        padding: 0.75em;
        text-align: left;
        border-bottom: 2px solid #dbdbdb;
        font-weight: 600;
        position: sticky;
        top: 0;
        background-color: #f5f5f5;
        z-index: 10;
      }

      #dataTable td {
        padding: 0.75em;
        border-bottom: 1px solid #dbdbdb;
        vertical-align: top;
      }

      #dataTable tbody tr {
        position: relative;
      }

      #dataTable tbody tr:hover {
        background-color: #fafafa;
      }

      #dataTable tbody tr td:first-child {
        position: relative;
        overflow: visible;
      }

      .row-link-icon {
        position: absolute;
        left: -24px;
        top: 50%;
        transform: translateY(-50%);
        opacity: 0.3;
        transition: opacity 0.2s;
        font-size: 1.1rem;
        z-index: 100;
      }

      #dataTable tbody tr:hover .row-link-icon {
        opacity: 1;
      }

      .row-link-icon a {
        text-decoration: none;
        color: #3273dc;
      }

      /* JSON wrapping */
      .json-content {
        white-space: pre-wrap !important;
        word-wrap: break-word !important;
        overflow-wrap: break-word !important;
        max-width: 100%;
      }

      pre.json-content {
        font-size: 0.75rem;
      }

      /* Pagination controls */
      .pagination-controls {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 1rem;
        margin-top: 1rem;
        padding: 1rem;
      }

      .pagination-controls button {
        padding: 0.5rem 1rem;
        border: 1px solid #dbdbdb;
        background-color: white;
        cursor: pointer;
        border-radius: 4px;
      }

      .pagination-controls button:hover:not(:disabled) {
        background-color: #f5f5f5;
      }

      .pagination-controls button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
      }

      .pagination-controls select {
        padding: 0.5rem;
        border: 1px solid #dbdbdb;
        border-radius: 4px;
      }
    </style>

  </head>

	<body>

		<section class="section">

    <div class="container" style="display: flex; justify-content: space-between;">
			{{ if .Name }}
			<h1 class="title">{{ .Name }}</h1>
			{{ end }}
			{{ if .Date }}
        <h2 class="subtitle" style="font-size: 2rem; margin-bottom: 0;">{{ formatDate .Date }}</h2>
        <a href="https://chalk.ai" target="_blank">
          <svg version="1.1" id="Layer_1" x="0px" y="0px" viewBox="0 0 577 194.68" xmlns="http://www.w3.org/2000/svg" style="width: 9rem" class="fill-black dark:fill-white h-8"><g><g><g><path d="M81.4,154.59c-24.04,0-39.42-19.96-39.42-42.53c0-22.08,14.88-42.53,38.77-42.53c19.14,0,33.53,10.47,36.64,29.77 H99.39c-1.96-10.47-8.67-16.52-18.32-16.52c-13.25,0-20.28,12.27-20.28,29.28c0,17.18,7.52,29.28,21.1,29.28 c9.49,0,16.36-6.38,18.16-16.85h18.16C115.1,143.8,100.54,154.59,81.4,154.59z"></path><path d="M134.24,152.96V40.09h17.34V87.2h0.33c3.43-11.45,11.94-17.67,23.88-17.67c17.99,0,26.83,12.27,26.83,30.26v53.16 h-17.83v-47.93c0-15.21-5.4-22.08-16.36-22.08c-10.96,0-16.36,7.52-16.36,20.94v49.07H134.24z"></path><path d="M247.6,154.59c-17.01,0-28.95-10.63-28.95-25.52c0-16.85,14.72-23.06,32.88-25.03l20.78-2.13v-2.45 c0-10.96-6.22-16.69-16.52-16.69c-9.98,0-15.86,5.07-16.36,14.23h-17.83c0.65-15.87,13.41-27.48,34.51-27.48 c20.78,0,34.02,11.12,34.02,33.21v50.22h-17.18v-16.03h-0.49C269.03,148.21,259.87,154.59,247.6,154.59z M252.18,141.34 c12.76,0,20.12-8.67,20.12-24.21v-2.78l-17.83,1.96c-10.8,1.14-17.01,5.07-17.01,12.92 C237.46,136.93,244.33,141.34,252.18,141.34z"></path><path d="M311.24,40.09h17.83v112.86h-17.83V40.09z"></path><path d="M368.81,110.26v42.69h-17.83V40.09h17.83v64.78l33.69-33.7h22.08l-35.82,35.99l37.46,45.8h-22.41L368.81,110.26z"></path></g></g><rect x="453.49" y="71.17" width="81.79" height="81.79"></rect></g></svg>
        </a>
			{{ end }}
		</div>

		</div>
		<br />

		<div class="container">
      <nav class="breadcrumb has-bullet-separator" aria-label="breadcrumbs">
        <ul>
          <li>
            <a href="#summary">
              <span class="icon is-small">
                <i class="fas fa-clipboard-list" aria-hidden="true"></i>
              </span>
              <span>Summary</span>
            </a>
          </li>
          <li>
            <a href="#histogram">
              <span class="icon is-small">
                <i class="fas fa-chart-bar" aria-hidden="true"></i>
              </span>
              <span>Histogram</span>
            </a>
          </li>
      	  <li>
            <a href="#rps">
              <span class="icon is-small">
                <i class="fas fa-chart-line" aria-hidden="true"></i>
              </span>
              <span> RPS </span>
            </a>
          </li>
          <li>
            <a href="#latency">
              <span class="icon is-small">
                <i class="far fa-clock" aria-hidden="true"></i>
              </span>
              <span>Latency Distribution</span>
            </a>
          </li>
          <li>
            <a href="#status">
              <span class="icon is-small">
                <i class="far fa-check-square" aria-hidden="true"></i>
              </span>
              <span>Status Distribution</span>
            </a>
					</li>
					{{ if gt (len .ErrorDist) 0 }}
          <li>
            <a href="#errors">
              <span class="icon is-small">
                <i class="fas fa-exclamation-circle" aria-hidden="true"></i>
              </span>
              <span>Errors</span>
            </a>
					</li>
					{{ end }}
          <li>
            <a href="#data">
              <span class="icon is-small">
                <i class="far fa-file-alt" aria-hidden="true"></i>
              </span>
              <span>Data</span>
            </a>
		  </li>
		  <li>
            <a href="#options">
              <span class="icon is-small">
                <i class="fas fa-cog" aria-hidden="true"></i>
              </span>
              <span>Options</span>
            </a>
          </li>
        </ul>
      </nav>
      <hr />
		</div>

		{{ if gt (len .Tags) 0 }}

			<div class="container">
				<div class="field is-grouped">

				{{ range $tag, $val := .Tags }}

					<div class="control">
						<div class="tags has-addons">
							<span class="tag is-dark">{{ $tag }}</span>
							<span class="tag is-primary">{{ $val }}</span>
						</div>
					</div>

				{{ end }}

				</div>
			</div>
			<br />
		{{ end }}

	  <div class="container">
			<div class="columns">
				<div class="column is-narrow">
					<div class="content">
						<a name="summary">
							<h3>Summary</h3>
						</a>
						<table class="table">
							<tbody>
								<tr>
									<th>Count</th>
									<td>{{ .Count }}</td>
								</tr>
								<tr>
									<th>Total</th>
									<td>{{ formatNanoUnit .Total }}</td>
								</tr>
								<tr>
									<th>Slowest</th>
								<td>{{ formatNanoUnit .Slowest }}</td>
								</tr>
								<tr>
									<th>Fastest</th>
									<td>{{ formatNanoUnit .Fastest }}</td>
								</tr>
								<tr>
									<th>Average</th>
									<td>{{ formatNanoUnit .Average }}</td>
								</tr>
								<tr>
									<th>Requests / sec</th>
									<td>{{ formatSeconds .Rps }}</td>
								</tr>
							</tbody>
						</table>
					</div>
				</div>
				<div class="column">
					<div class="content">
					</div>
				</div>
			</div>
	  </div>

	  <br />
		<div class="container">
			<div class="content">
				<a name="histogram">
					<h3>Histogram</h3>
				</a>
				<p>
					<canvas id="js-bar-container"></>
				</p>
			</div>
	  </div>

	  <br />
	  <div class="container">
	    <hr style="border: 0; height: 2px; background-color: #dbdbdb; margin: 2rem 0;" />
	  </div>

		<div class="container">
			<div class="content">
				<a name="rps">
					<h3> RPS </h3>
				</a>
				<p>
					<canvas id="js-rps-container"></canvas>
				</p>
			</div>
	  </div>

	  <br />
		<div class="container">
			<div class="content">
				<a name="latency">
					<h3>Latency distribution</h3>
				</a>
				<table class="table is-fullwidth">
					<thead>
						<tr>
							{{ range .LatencyDistribution }}
								<th>{{ .Percentage }} %</th>
							{{ end }}
						</tr>
					</thead>
					<tbody>
						<tr>
							{{ range .LatencyDistribution }}
								<td>{{ formatNanoUnit .Latency }}</td>
							{{ end }}
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		<br />
		<div class="container">
			<div class="columns">
				<div class="column is-narrow">
					<div class="content">
						<a name="status">
							<h3>Status distribution</h3>
						</a>
						<table class="table is-hoverable">
							<thead>
								<tr>
									<th>Status</th>
									<th>Count</th>
									<th>% of Total</th>
								</tr>
							</thead>
							<tbody>
							  {{ range $code, $num := .StatusCodeDist }}
									<tr>
									  <td>{{ $code }}</td>
										<td>{{ $num }}</td>
										<td>{{ formatPercent $num $.Count }} %</td>
									</tr>
									{{ end }}
								</tbody>
							</table>
						</div>
					</div>
				</div>
			</div>

			{{ if gt (len .ErrorDist) 0 }}

				<br />
				<div class="container">
					<div class="columns">
						<div class="column is-narrow">
							<div class="content">
								<a name="errors">
									<h3>Errors</h3>
								</a>
								<table class="table is-hoverable">
									<thead>
										<tr>
											<th>Error</th>
											<th>Count</th>
											<th>% of Total</th>
										</tr>
									</thead>
									<tbody>
										{{ range $err, $num := .ErrorDist }}
											<tr>
												<td>{{ $err }}</td>
												<td>{{ $num }}</td>
												<td>{{ formatPercent $num $.Count }} %</td>
											</tr>
											{{ end }}
										</tbody>
									</table>
								</div>
							</div>
						</div>
					</div>

			{{ end }}

			<br />
      <div class="container">
        <div class="content">
          <a name="data">
            <h3>Data</h3>
          </a>

          <article class="message is-info">
            <div class="message-body">
              {{ if gt .Options.DataSamplingRate 0.0 }}
              <p><strong>Sample:</strong> Showing {{ len .Details }} of {{ .Count }} requests ({{ printf "%.2f%%" .Options.DataSamplingRate }} sampling rate)</p>
              <p><strong>Note:</strong> Request/response payloads included for sampled requests</p>
              {{ else }}
              <p><strong>Sample:</strong> No data sampling configured</p>
              {{ end }}
            </div>
          </article>

        </div>

        <div id="dataTableContainer" style="margin-top: 20px;">
          <div class="field" style="margin-bottom: 12px;">
            <label class="switch" style="margin-right: 8px;">
              <input type="checkbox" id="globalJsonToggle" onchange="toggleAllJson()">
              <span class="switch-slider"></span>
            </label>
            <span style="font-size: 0.85rem; color: #363636;">Show JSON</span>
          </div>

          <div style="margin-left: 35px; overflow: visible; position: relative;">
            <div id="dataTable"></div>

            <div class="pagination-controls">
              <button id="firstPageBtn" onclick="goToFirstPage()">First</button>
              <button id="prevPageBtn" onclick="goToPrevPage()">Previous</button>
              <span>Page <strong id="currentPageDisplay">1</strong> of <strong id="totalPagesDisplay">1</strong></span>
              <button id="nextPageBtn" onclick="goToNextPage()">Next</button>
              <button id="lastPageBtn" onclick="goToLastPage()">Last</button>
              <select id="pageSizeSelect" onchange="changePageSize()">
                <option value="10">10 per page</option>
                <option value="25" selected>25 per page</option>
                <option value="50">50 per page</option>
                <option value="100">100 per page</option>
              </select>
            </div>
          </div>
        </div>
			</div>

			<br />

			<div class="container">
				<div class="content">
					<a name="options">
						<h3>Options</h3>
					</a>
					<article class="message">
						<div class="message-body">
							<pre style="background-color: transparent;">{{ jsonify .Options true }}</pre>
						</div>
					</article>
				</div>
			</div>

			<div class="container">
        <hr />
        <div class="content has-text-centered">
          <p>
            Generated by <strong>ghz</strong>
          </p>
          <a href="https://github.com/bojand/ghz"><i class="icon is-medium fab fa-github"></i></a>
        </div>
      </div>

		</section>

  </body>

  <script>
	const count = {{ .Count }};

	const rawData = {{ jsonify .Details false }};
	const sampleCount = {{ len .Details }};
	const totalCount = {{ .Count }};

	const data = [
		{{ range .Histogram }}
			{ name: "{{ .AlternativeMark }}", value: {{ .Count }} },
		{{ end }}
	];

	const rps = {{ jsonify .RPS false }}

	const aggData = {{ jsonify .Aggs false }}

	const createBarChart = () => {
	  const ctx = document.getElementById('js-bar-container');

	  new Chart(ctx, {
		type: 'bar',
		data: {
		  labels: data.map((item) => item.name),
		  datasets: [{
			label: 'Latency Bucket',
			data: data.map((item) => item.value),
			borderWidth: 1,
			barThickness: 20,
			backgroundColor: ["#B8CEC6", "#7BA392", "#3F7067", "#32645B", "#29524A", "#264A43", "#1C3B34", "#0C3129", "#06261F", "#041D17"]
		  }]
		},
		options: {
		  aspectRatio: 3,
		  plugins: {
			legend: {
			  display: false
			},
		  },
		  tooltips: {
			callbacks: {
			  label: function(tooltipItem) {
				return tooltipItem.yLabel;
			  }
			}
		  },
		  indexAxis: 'y',
		  scales: {
			y: {
			  ticks: {
				callback: function(value) {
				  if (value == data.length - 1) {
					return data[value].name.split(' ')[0]
				  } else {
					return parseFloat(data[value].name).toFixed(1) + 'ms'
				  }
				}
			  }
			}
		  }
		},
	  });
	}

	const createRPSChart = () => {
	  const ctx = document.getElementById('js-rps-container');

	  new Chart(ctx, {
		type: 'line',
		data: {
		  labels: rps.map((item) => (item.x / 1000)),
		  datasets: [
			{
			  label: 'RPS',
			  data: rps,
			  borderWidth: 2,
			  borderColor: "rgba(38, 74, 67, 1)",
			  backgroundColor: "rgba(38, 74, 67, 1)",
			  pointRadius: 0,
			  tension: 0.3,
			},
			{{ if .P50}}
			{
			  label: 'P50',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p50 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(112, 206, 151)",
			  backgroundColor: "rgb(112, 206, 151, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  lineWidth: 2,
			  yAxisID: "y2"
			},
			{{ end }}
			{{ if .P95}}
			{
			  label: 'P95',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p95 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(129, 206, 234)",
			  backgroundColor: "rgb(129, 206, 234, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  lineWidth: 2,
			  yAxisID: "y2"
			},
			{{ end }}
			{{ if .P99 }}
			{
			  label: 'P99',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p99 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(249, 149, 127)",
			  backgroundColor: "rgb(249, 149, 127, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  yAxisID: "y2"
			},
			{{ end }}
			{{ if .P99_9 }}
			{
			  label: 'P99.9',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p99_9 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(220, 100, 100)",
			  backgroundColor: "rgb(220, 100, 100, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  yAxisID: "y2"
			}
			{{ end }}
		  ]
		},
		options: {
		  plugins: {
			legend: {
				align: "end",
				labels: {
				  boxWidth: 20
				}
			  }
		  },
		  aspectRatio: 2.5,
		  layout: {
			padding: {
			  left: 30
			}
		  },
		  scales: {
			y:
			{
			  beginAtZero: true,
			  type: 'linear',
			  position: 'left',
			  stack: "demo",
			  stackWeight: 2,
			},
			y2: {
			  type: 'linear',
			  position: 'left',
			  offset: true,
			  stack: "demo",
			  stackWeight: 3,
			  ticks: {
          callback: function(value) {
            return value + 'ms'
          },
			  },
			},
			x: {
          ticks: {
            callback: function(value) {
              return value < 60 ? value + 's' : Math.floor(value / 60) + 'm' + value % 60 + 's'
            },
            autoSkip: true,
            maxTicksLimit: 20,
          }
        }
		  }
		}
	  });
	}

	// Table state
	const tableData = rawData || [];
	const hasPayloads = {{ (gt .Options.DataSamplingRate 0.0) }};
	let pageIndex = 0;
	let pageSize = 25;

	function loadAndShowTable() {
	  const btn = document.getElementById('showTableBtn');
	  if (btn) {
		btn.classList.add('is-loading');
		btn.disabled = true;
	  }

	  try {
		console.log('Loading data...');
		console.log('Sample data row:', tableData.length > 0 ? tableData[0] : 'No data');
		console.log('Capture payloads:', hasPayloads);

		if (tableData.length === 0) {
		  alert('No data found in sample. This might indicate an issue with data collection.');
		  if (btn) {
			btn.classList.remove('is-loading');
			btn.disabled = false;
		  }
		  return;
		}

		renderTable();
		document.getElementById('dataTableContainer').style.display = 'block';
		if (btn) btn.style.display = 'none';
	  } catch (error) {
		console.error('Error loading table:', error);
		if (btn) {
		  btn.classList.remove('is-loading');
		  btn.disabled = false;
		}
		alert('Failed to load table data: ' + error.message);
	  }
	}

	function changePageSize() {
	  pageSize = parseInt(document.getElementById('pageSizeSelect').value);
	  pageIndex = 0;
	  renderTable();
	}

	function goToFirstPage() {
	  pageIndex = 0;
	  renderTable();
	}

	function goToPrevPage() {
	  if (pageIndex > 0) {
		pageIndex--;
		renderTable();
	  }
	}

	function goToNextPage() {
	  const totalPages = Math.ceil(tableData.length / pageSize);
	  if (pageIndex < totalPages - 1) {
		pageIndex++;
		renderTable();
	  }
	}

	function goToLastPage() {
	  const totalPages = Math.ceil(tableData.length / pageSize);
	  pageIndex = totalPages - 1;
	  renderTable();
	}

	function formatLatency(latency) {
	  if (latency !== undefined && latency !== null) {
		// latency is in nanoseconds
		const ms = latency / 1000000;
		return ms.toFixed(2) + ' ms';
	  }
	  return '-';
	}

	function formatTimestamp(timestamp) {
	  if (timestamp) {
		const d = new Date(timestamp);
		return d.toLocaleString();
	  }
	  return '-';
	}

	function syntaxHighlight(json) {
	  json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
	  return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
		var cls = 'json-number';
		if (/^"/.test(match)) {
		  if (/:$/.test(match)) {
			cls = 'json-key';
		  } else {
			cls = 'json-string';
		  }
		} else if (/true|false/.test(match)) {
		  cls = 'json-boolean';
		} else if (/null/.test(match)) {
		  cls = 'json-null';
		}
		return '<span class="' + cls + '">' + match + '</span>';
	  });
	}

	function decodeFeatherToTable(base64Data) {
	  try {
		// Check if Arrow library is loaded
		if (typeof Arrow === 'undefined') {
		  return '<div style="color: red; font-size: 0.7rem;">Apache Arrow library not loaded</div>';
		}

		// Decode base64
		const binaryString = atob(base64Data);
		const bytes = new Uint8Array(binaryString.length);
		for (let i = 0; i < binaryString.length; i++) {
		  bytes[i] = binaryString.charCodeAt(i);
		}

		// Parse Arrow IPC format (Feather)
		const table = Arrow.tableFromIPC(bytes);

		// Convert to transposed HTML table (columns as rows)
		let html = '<table class="table is-narrow is-striped" style="font-size: 0.75rem; margin: 0; border: 1px solid #dbdbdb; border-radius: 4px;"><tbody>';

		// Limit to first 10 rows
		const maxRows = Math.min(10, table.numRows);

		// Each column becomes a row in the transposed table
		for (let j = 0; j < table.numCols; j++) {
		  const field = table.schema.fields[j];
		  html += '<tr>';
		  html += '<th style="background-color: #f5f5f5; font-weight: 600; white-space: nowrap; padding: 4px 8px;">' + field.name + '</th>';

		  // Show values for this column across all rows
		  for (let i = 0; i < maxRows; i++) {
			const val = table.getChildAt(j).get(i);
			const displayVal = val !== null ? String(val) : '<span style="color: #aaa;">null</span>';
			html += '<td style="padding: 4px 8px;">' + displayVal + '</td>';
		  }

		  if (table.numRows > 10) {
			html += '<td style="padding: 4px 8px; font-style: italic; color: #888;">...</td>';
		  }

		  html += '</tr>';
		}

		html += '</tbody></table>';
		html += '<div style="font-size: 0.65rem; color: #888; margin-top: 4px;">' + table.numRows + ' rows × ' + table.numCols + ' cols</div>';

		return html;
	  } catch (e) {
		return '<div style="color: #856404; background-color: #fff3cd; border: 1px solid #ffeeba; padding: 8px; border-radius: 4px; font-size: 0.75rem; margin: 4px 0;">' +
		  '<strong>⚠️ Warning:</strong> Failed to parse Feather/Arrow data.<br>' +
		  '<strong>Error:</strong> ' + e.message + '<br>' +
		  '<strong>Note:</strong> Arrow.js may have issues with large lists. Consider using JSON format instead for better compatibility.' +
		  '</div>';
	  }
	}

	function formatPayload(payload, isRequest) {
	  if (!payload) return { link: null, html: '-' };

	  try {
		// Try to parse as JSON (protobuf messages are serialized as JSON)
		const parsed = JSON.parse(payload);
		const formatted = JSON.stringify(parsed, null, 2);

		// Check for feather data in response (scalars_data) or request (inputs/inputs_feather)
		let featherHtml = null;
		let featherField = null;
		let queryInfo = null;

		if (parsed.scalars_data && typeof parsed.scalars_data === 'string') {
		  featherField = 'scalars_data';
		  featherHtml = decodeFeatherToTable(parsed.scalars_data);

		  // Extract query metadata if available
		  if (parsed.response_meta) {
			queryInfo = {
			  query_id: parsed.response_meta.query_id,
			  environment_id: parsed.response_meta.environment_id,
			  query_timestamp: parsed.response_meta.query_timestamp,
			  execution_duration: parsed.response_meta.execution_duration
			};
		  }
		} else if (parsed.inputs_feather && typeof parsed.inputs_feather === 'string') {
		  featherField = 'inputs_feather';
		  featherHtml = decodeFeatherToTable(parsed.inputs_feather);
		} else if (parsed.inputs && typeof parsed.inputs === 'string') {
		  featherField = 'inputs';
		  featherHtml = decodeFeatherToTable(parsed.inputs);
		}

		// Create a compact summary for protobuf messages
		const summary = createPayloadSummary(parsed);

		// Build link HTML if we have query info
		let linkHtml = null;
		if (queryInfo && queryInfo.query_id && queryInfo.environment_id) {
		  const timestamp = new Date(queryInfo.query_timestamp).getTime();
		  const chalkUrl = 'https://chalk.ai/projects/inacjutizsafg/environments/' + queryInfo.environment_id + '/query-runs/' + queryInfo.query_id + '?ts=' + timestamp;
		  linkHtml = '<a href="' + chalkUrl + '" target="_blank" style="color: #3273dc; font-size: 1rem;" title="View in Chalk (duration: ' + (queryInfo.execution_duration || 'N/A') + ')">🔗</a>';
		}

		// If it's very short and no feather data, just show it all with highlighting
		if (formatted.length <= 100 && !featherHtml) {
		  return {
			link: linkHtml,
			html: '<pre class="json-content" style="margin: 0; font-size: 0.75rem;">' + syntaxHighlight(formatted) + '</pre>'
		  };
		}

		// Generate unique ID for this payload
		const id = 'payload-' + Math.random().toString(36).substring(7);

		let html = '<div>';

		// If we have feather data, show table by default controlled by global toggle
		if (featherHtml) {
		  html += '<div class="payload-table-view" style="margin: 0; overflow-x: auto; display: block;">' + featherHtml + '</div>';
		  html += '<pre class="payload-json-view json-content" style="margin: 0; font-size: 0.75rem; display: none;">' + syntaxHighlight(formatted) + '</pre>';
		} else {
		  // No feather data, show summary/expand as before
		  html += '<div id="' + id + '-summary" style="margin: 0; display: block;">' + summary + '</div>';
		  html += '<pre id="' + id + '-full" class="json-content" style="margin: 0; font-size: 0.75rem; display: none;">' + syntaxHighlight(formatted) + '</pre>';
		  html += '<a href="#" onclick="togglePayload(\'' + id + '\'); return false;" style="font-size: 0.75rem; color: #3273dc;">Expand</a>';
		}

		html += '</div>';
		return { link: linkHtml, html: html };
	  } catch (e) {
		// Not valid JSON, show as plain text
		const text = payload.length <= 100 ? payload : payload.substring(0, 100) + '...';
		return { link: null, html: text };
	  }
	}

	function createPayloadSummary(obj, maxDepth = 2) {
	  if (typeof obj !== 'object' || obj === null) {
		return String(obj);
	  }

	  if (Array.isArray(obj)) {
		return '[' + obj.length + ' items]';
	  }

	  // Create a compact representation of the object
	  const keys = Object.keys(obj);
	  if (keys.length === 0) return '{}';

	  const parts = [];
	  for (let i = 0; i < Math.min(keys.length, 3); i++) {
		const key = keys[i];
		const val = obj[key];

		if (typeof val === 'object' && val !== null) {
		  if (Array.isArray(val)) {
			parts.push(key + ': [' + val.length + ']');
		  } else {
			parts.push(key + ': {...}');
		  }
		} else {
		  const valStr = String(val);
		  parts.push(key + ': ' + (valStr.length > 20 ? valStr.substring(0, 20) + '...' : valStr));
		}
	  }

	  let result = '{ ' + parts.join(', ');
	  if (keys.length > 3) {
		result += ', ... +' + (keys.length - 3) + ' fields';
	  }
	  result += ' }';

	  return result;
	}

	function togglePayload(id) {
	  const summary = document.getElementById(id + '-summary');
	  const full = document.getElementById(id + '-full');
	  const link = event.target;

	  if (summary.style.display === 'none') {
		summary.style.display = 'block';
		full.style.display = 'none';
		link.textContent = 'Expand';
	  } else {
		summary.style.display = 'none';
		full.style.display = 'block';
		link.textContent = 'Collapse';
	  }
	}

	function toggleAllJson() {
	  const toggle = document.getElementById('globalJsonToggle');
	  const isChecked = toggle.checked;

	  // Find all payload table and json views
	  const tableViews = document.querySelectorAll('.payload-table-view');
	  const jsonViews = document.querySelectorAll('.payload-json-view');

	  if (isChecked) {
		// Show JSON
		tableViews.forEach(el => el.style.display = 'none');
		jsonViews.forEach(el => el.style.display = 'block');
	  } else {
		// Show tables
		tableViews.forEach(el => el.style.display = 'block');
		jsonViews.forEach(el => el.style.display = 'none');
	  }
	}

	function renderTable() {
	  const container = document.getElementById('dataTable');
	  if (!container) {
		console.error('dataTable container not found!');
		return;
	  }

	  // Calculate pagination
	  const startIdx = pageIndex * pageSize;
	  const endIdx = Math.min(startIdx + pageSize, tableData.length);
	  const pageData = tableData.slice(startIdx, endIdx);
	  const totalPages = Math.ceil(tableData.length / pageSize);

	  // Build table HTML
	  let html = '<table><thead><tr>';

	  // Headers
	  html += '<th>Timestamp</th>';
	  html += '<th>Latency</th>';
	  html += '<th>Status</th>';
	  html += '<th>Error</th>';
	  if (hasPayloads) {
		html += '<th>Request</th>';
		html += '<th>Response</th>';
	  }
	  html += '</tr></thead><tbody>';

	  // Rows
	  pageData.forEach(row => {
		// Format payloads once and reuse
		let requestFormatted = null;
		let responseFormatted = null;
		let rowLinkIcon = '';

		if (hasPayloads) {
		  try {
			requestFormatted = formatPayload(row.requestPayload, true);
			responseFormatted = formatPayload(row.responsePayload, false);

			// Get link from response or request
			const linkHtml = (responseFormatted && responseFormatted.link) || (requestFormatted && requestFormatted.link);
			if (linkHtml) {
			  rowLinkIcon = '<span class="row-link-icon">' + linkHtml + '</span>';
			}
		  } catch (err) {
			console.error('Error formatting payload:', err);
		  }
		}

		html += '<tr>';
		html += '<td style="position: relative;">' + rowLinkIcon + formatTimestamp(row.timestamp) + '</td>';
		html += '<td>' + formatLatency(row.latency) + '</td>';
		html += '<td><span class="tag ' + (row.status === 'OK' ? 'is-success' : 'is-danger') + '">' + (row.status || '-') + '</span></td>';
		html += '<td>' + (row.error || '-') + '</td>';

		if (hasPayloads) {
		  html += '<td>' + (requestFormatted ? requestFormatted.html : '-') + '</td>';
		  html += '<td>' + (responseFormatted ? responseFormatted.html : '-') + '</td>';
		}

		html += '</tr>';
	  });

	  html += '</tbody></table>';
	  container.innerHTML = html;

	  // Update pagination controls
	  const currentPageEl = document.getElementById('currentPageDisplay');
	  const totalPagesEl = document.getElementById('totalPagesDisplay');
	  const firstBtn = document.getElementById('firstPageBtn');
	  const prevBtn = document.getElementById('prevPageBtn');
	  const nextBtn = document.getElementById('nextPageBtn');
	  const lastBtn = document.getElementById('lastPageBtn');

	  if (currentPageEl) currentPageEl.textContent = pageIndex + 1;
	  if (totalPagesEl) totalPagesEl.textContent = totalPages;
	  if (firstBtn) firstBtn.disabled = pageIndex === 0;
	  if (prevBtn) prevBtn.disabled = pageIndex === 0;
	  if (nextBtn) nextBtn.disabled = pageIndex >= totalPages - 1;
	  if (lastBtn) lastBtn.disabled = pageIndex >= totalPages - 1;
	}

	createBarChart();

	createRPSChart();

	// Auto-load first page of table data when DOM is ready
	function initTableOnLoad() {
	  console.log('DOM ready, initializing table...');
	  console.log('tableData length:', tableData ? tableData.length : 'undefined');
	  console.log('hasPayloads:', hasPayloads);

	  if (tableData && tableData.length > 0) {
		try {
		  console.log('Rendering table...');
		  renderTable();
		  console.log('Table rendered successfully');
		} catch (error) {
		  console.error('Error rendering table:', error);
		  console.error('Error stack:', error.stack);
		}
	  } else {
		console.log('No table data to display');
	  }
	}

	// Wait for DOM to be ready
	if (document.readyState === 'loading') {
	  document.addEventListener('DOMContentLoaded', initTableOnLoad);
	} else {
	  // DOM already loaded
	  initTableOnLoad();
	}
	</script>
	<script defer src="https://use.fontawesome.com/releases/v5.1.0/js/all.js"></script>
</html>
`
)
