const COLORS = {
  accent: "FF65727D",
  dark: "FF3F4852",
  muted: "FF7A838C",
  line: "FFD7DBDF",
  light: "FFF5F6F7",
  success: "FFEAF0E8",
  successText: "FF4D6848",
  danger: "FFF3E9E9",
  dangerText: "FF8A4B4B",
};

const MONEY_FORMAT = '#,##0.00 "₽"';
const THIN_BORDER = {
  top: { style: "thin", color: { argb: COLORS.line } },
  left: { style: "thin", color: { argb: COLORS.line } },
  bottom: { style: "thin", color: { argb: COLORS.line } },
  right: { style: "thin", color: { argb: COLORS.line } },
};

const VERDICT_LABELS = {
  PASS: "Критерий выполняется",
  FAIL: "Критерий не выполняется",
  NO_THRESHOLD: "Планка не задана",
};

const THRESHOLD_LABELS = {
  RF: "По РФ",
  REGION: "По региону",
};

const RECOMMENDATION_LABELS = {
  CAN_DECREASE: "Допустимый запас на сотрудника в месяц",
  NEED_INCREASE: "Необходимое увеличение на сотрудника в месяц",
  NO_THRESHOLD: "Планка для сравнения не задана",
};

function toNumber(value){
  if (value === null || value === undefined || value === ""){
    return null;
  }

  const number = Number(String(value).replace(/\s/g, "").replace(",", "."));
  return Number.isFinite(number) ? number : value;
}

function numeric(value){
  const number = toNumber(value);
  return typeof number === "number" ? number : 0;
}

function displayNumber(value){
  const number = toNumber(value);
  return typeof number === "number" ? number.toLocaleString("ru-RU") : "—";
}

function getMonth(employee, monthIndex){
  return employee?.months?.find((month) => Number(month.month_index) === monthIndex)
    ?? employee?.months?.[monthIndex - 1]
    ?? {};
}

function getMonthResult(result, monthIndex){
  return result?.monthly_results?.find((month) => Number(month.month_index) === monthIndex)
    ?? result?.monthly_results?.[monthIndex - 1]
    ?? {};
}

function eachCell(sheet, startRow, startColumn, endRow, endColumn, callback){
  for (let row = startRow; row <= endRow; row += 1){
    for (let column = startColumn; column <= endColumn; column += 1){
      callback(sheet.getCell(row, column));
    }
  }
}

function styleCells(sheet, startRow, startColumn, endRow, endColumn, style){
  eachCell(sheet, startRow, startColumn, endRow, endColumn, (cell) =>
    Object.assign(cell, style)
  );
}

function addSection(sheet, row, title){
  sheet.mergeCells(row, 1, row, 8);
  const cell = sheet.getCell(row, 1);
  cell.value = title;
  cell.fill = { type: "pattern", pattern: "solid", fgColor: { argb: COLORS.dark } };
  cell.font = { name: "Arial", size: 11, bold: true, color: { argb: "FFFFFFFF" } };
  cell.alignment = { vertical: "middle" };
  sheet.getRow(row).height = 24;
  return row + 1;
}

function addInfoRow(sheet, row, label, value, numberFormat){
  sheet.mergeCells(row, 1, row, 2);
  sheet.mergeCells(row, 3, row, 8);
  sheet.getCell(row, 1).value = label;
  sheet.getCell(row, 3).value = value ?? "—";
  sheet.getCell(row, 1).font = { name: "Arial", size: 10, bold: true, color: { argb: COLORS.muted } };
  sheet.getCell(row, 3).font = { name: "Arial", size: 10, color: { argb: COLORS.dark } };
  sheet.getCell(row, 3).numFmt = numberFormat ?? "General";
  styleCells(sheet, row, 1, row, 8, {
    alignment: { vertical: "middle", wrapText: true },
    border: THIN_BORDER,
  });
  sheet.getRow(row).height = 21;
  return row + 1;
}

function addTableHeader(sheet, row, values){
  const cells = sheet.getRow(row);
  values.forEach((value, index) => {
    const cell = cells.getCell(index + 1);
    cell.value = value;
    cell.fill = { type: "pattern", pattern: "solid", fgColor: { argb: COLORS.accent } };
    cell.font = { name: "Arial", size: 9, bold: true, color: { argb: "FFFFFFFF" } };
    cell.alignment = { horizontal: "center", vertical: "middle", wrapText: true };
    cell.border = THIN_BORDER;
  });
  cells.height = 30;
}

function formula(formula, result){
  return { formula, result };
}

function sanitizeFileName(value){
  const safe = String(value ?? "отчёт")
    .trim()
    .replace(/[<>:"/\\|?*\u0000-\u001F]/g, "_")
    .replace(/\s+/g, " ")
    .slice(0, 70);
  return safe || "отчёт";
}

function buildReportSheet(workbook, calculation, monthNames){
  const input = calculation.input ?? {};
  const result = calculation.result ?? {};
  const employees = input.employees ?? [];
  const period = input.period ?? {};
  const thresholdInput = input.threshold ?? {};
  const thresholdResult = result.threshold ?? {};
  const quarterResult = result.quarter ?? {};
  const comparison = result.comparison ?? {};
  const recommendation = result.recommendation ?? {};
  const sheet = workbook.addWorksheet("Отчёт", {
    pageSetup: {
      paperSize: 9,
      orientation: "landscape",
      fitToPage: true,
      fitToWidth: 1,
      fitToHeight: 0,
      horizontalCentered: true,
      margins: { left: 0.3, right: 0.3, top: 0.45, bottom: 0.45, header: 0.2, footer: 0.2 },
    },
    views: [{ showGridLines: false, zoomScale: 85 }],
  });

  sheet.columns = [
    { key: "employee", width: 10 },
    { key: "m1total", width: 17 },
    { key: "m1gph", width: 15 },
    { key: "m2total", width: 17 },
    { key: "m2gph", width: 15 },
    { key: "m3total", width: 17 },
    { key: "m3gph", width: 15 },
    { key: "labor", width: 24 },
  ];
  sheet.properties.defaultRowHeight = 19;

  sheet.mergeCells("A1:H1");
  sheet.getCell("A1").value = input.title ?? "Отчёт по выплатам";
  sheet.getCell("A1").font = { name: "Arial", size: 18, bold: true, color: { argb: "FFFFFFFF" } };
  sheet.getCell("A1").fill = { type: "pattern", pattern: "solid", fgColor: { argb: COLORS.accent } };
  sheet.getCell("A1").alignment = { vertical: "middle" };
  sheet.getRow(1).height = 36;

  sheet.mergeCells("A2:H2");
  sheet.getCell("A2").value = "Расчёт среднемесячного размера выплат по организации";
  sheet.getCell("A2").font = { name: "Arial", size: 11, italic: true, color: { argb: COLORS.muted } };
  sheet.getCell("A2").alignment = { vertical: "middle" };
  sheet.getRow(2).height = 25;

  let row = addSection(sheet, 4, "1. Реквизиты отчёта");
  row = addInfoRow(sheet, row, "Организация", input.org_name);
  row = addInfoRow(sheet, row, "Отчётный период", `${period.quarter ?? "—"} квартал ${period.year ?? "—"} года`);
  if (calculation.id !== undefined){
    row = addInfoRow(sheet, row, "Идентификатор расчёта", calculation.id);
  }
  if (calculation.checked_at ?? calculation.created_at){
    const date = new Date(calculation.checked_at ?? calculation.created_at);
    row = addInfoRow(
      sheet,
      row,
      "Дата расчёта",
      Number.isNaN(date.getTime()) ? "—" : new Intl.DateTimeFormat("ru-RU", {
        dateStyle: "short",
        timeStyle: "short",
      }).format(date)
    );
  }
  row += 1;

  row = addSection(sheet, row, "2. Исходные планки Росстата");
  row = addInfoRow(sheet, row, "По РФ", toNumber(thresholdInput.value_rf), MONEY_FORMAT);
  row = addInfoRow(sheet, row, "По региону", toNumber(thresholdInput.value_region), MONEY_FORMAT);
  row = addInfoRow(sheet, row, "Код региона", thresholdInput.region_code);
  row += 1;

  row = addSection(sheet, row, "3. Исходные выплаты по сотрудникам");
  const employeeHeaderRow = row;
  sheet.mergeCells(row, 1, row + 1, 1);
  sheet.getCell(row, 1).value = "Сотрудник";
  monthNames.slice(0, 3).forEach((name, index) => {
    const startColumn = 2 + index * 2;
    sheet.mergeCells(row, startColumn, row, startColumn + 1);
    sheet.getCell(row, startColumn).value = name;
  });
  sheet.mergeCells(row, 8, row + 1, 8);
  sheet.getCell(row, 8).value = "Выплаты по ТД за квартал";
  [1, 2, 4, 6, 8].forEach((column) => {
    const cell = sheet.getCell(row, column);
    cell.fill = { type: "pattern", pattern: "solid", fgColor: { argb: COLORS.accent } };
    cell.font = { name: "Arial", size: 9, bold: true, color: { argb: "FFFFFFFF" } };
    cell.alignment = { horizontal: "center", vertical: "middle", wrapText: true };
  });
  [2, 3, 4, 5, 6, 7].forEach((column) => {
    const cell = sheet.getCell(row + 1, column);
    cell.value = column % 2 === 0 ? "Всего (стр. 140)" : "ГПХ (стр. 160)";
    cell.fill = { type: "pattern", pattern: "solid", fgColor: { argb: COLORS.accent } };
    cell.font = { name: "Arial", size: 8, bold: true, color: { argb: "FFFFFFFF" } };
    cell.alignment = { horizontal: "center", vertical: "middle", wrapText: true };
  });
  styleCells(sheet, row, 1, row + 1, 8, { border: THIN_BORDER });
  sheet.getRow(row).height = 25;
  sheet.getRow(row + 1).height = 30;
  row += 2;

  const employeeStartRow = row;
  employees.forEach((employee) => {
    const values = [employee.employee_index];
    let laborTotal = 0;
    for (let monthIndex = 1; monthIndex <= 3; monthIndex += 1){
      const month = getMonth(employee, monthIndex);
      const total = toNumber(month.total_140);
      const gph = toNumber(month.gph_160);
      values.push(total, gph);
      laborTotal += numeric(total) - numeric(gph);
    }
    values.push(formula(`SUM(B${row},D${row},F${row})-SUM(C${row},E${row},G${row})`, laborTotal));
    sheet.getRow(row).values = values;
    styleCells(sheet, row, 1, row, 8, {
      font: { name: "Arial", size: 9, color: { argb: COLORS.dark } },
      alignment: { horizontal: "right", vertical: "middle" },
      border: THIN_BORDER,
    });
    sheet.getCell(row, 1).alignment = { horizontal: "center", vertical: "middle" };
    eachCell(sheet, row, 2, row, 8, (cell) => { cell.numFmt = MONEY_FORMAT; });
    row += 1;
  });

  const employeeEndRow = row - 1;
  const totals = ["Итого"];
  for (let column = 2; column <= 8; column += 1){
    const letter = sheet.getColumn(column).letter;
    const resultValue = employees.reduce((sum, employee) => {
      if (column === 8){
        return sum + [1, 2, 3].reduce((employeeSum, monthIndex) => {
          const month = getMonth(employee, monthIndex);
          return employeeSum + numeric(month.total_140) - numeric(month.gph_160);
        }, 0);
      }
      const monthIndex = Math.ceil((column - 1) / 2);
      const field = column % 2 === 0 ? "total_140" : "gph_160";
      return sum + numeric(getMonth(employee, monthIndex)[field]);
    }, 0);
    totals.push(employeeEndRow >= employeeStartRow
      ? formula(`SUM(${letter}${employeeStartRow}:${letter}${employeeEndRow})`, resultValue)
      : 0);
  }
  sheet.getRow(row).values = totals;
  styleCells(sheet, row, 1, row, 8, {
    fill: { type: "pattern", pattern: "solid", fgColor: { argb: COLORS.light } },
    font: { name: "Arial", size: 9, bold: true, color: { argb: COLORS.dark } },
    alignment: { horizontal: "right", vertical: "middle" },
    border: THIN_BORDER,
  });
  eachCell(sheet, row, 2, row, 8, (cell) => { cell.numFmt = MONEY_FORMAT; });
  row += 2;

  row = addSection(sheet, row, "4. Лесенка расчёта");
  sheet.mergeCells(row, 1, row, 8);
  sheet.getCell(row, 1).value = "Шаг 1. По каждому сотруднику: выплаты по ТД = всего (стр. 140) − ГПХ (стр. 160). Итоги показаны в последнем столбце таблицы выше.";
  sheet.getCell(row, 1).alignment = { wrapText: true, vertical: "middle" };
  sheet.getCell(row, 1).font = { name: "Arial", size: 9, color: { argb: COLORS.dark } };
  sheet.getCell(row, 1).border = THIN_BORDER;
  sheet.getRow(row).height = 32;
  row += 1;

  sheet.mergeCells(row, 1, row, 8);
  sheet.getCell(row, 1).value = "Шаг 2. Средняя выплата за каждый месяц";
  sheet.getCell(row, 1).font = { name: "Arial", size: 10, bold: true, color: { argb: COLORS.dark } };
  sheet.getCell(row, 1).alignment = { vertical: "middle" };
  row += 1;
  addTableHeader(sheet, row, ["Месяц", "Выплаты по ТД", "Физлица", "Средняя", "Формула", "", "", ""]);
  sheet.mergeCells(row, 5, row, 8);
  row += 1;
  const monthlyStartRow = row;
  for (let monthIndex = 1; monthIndex <= 3; monthIndex += 1){
    const monthResult = getMonthResult(result, monthIndex);
    const laborSum = numeric(monthResult.labor_payment_sum);
    const headcount = numeric(monthResult.headcount);
    const average = numeric(monthResult.average);
    sheet.getRow(row).values = [
      monthNames[monthIndex - 1] ?? `Месяц ${monthIndex}`,
      toNumber(monthResult.labor_payment_sum),
      toNumber(monthResult.headcount),
      formula(`IFERROR(B${row}/C${row},0)`, average),
      `${laborSum.toLocaleString("ru-RU")} ÷ ${headcount} = ${average.toLocaleString("ru-RU")}`,
    ];
    sheet.mergeCells(row, 5, row, 8);
    styleCells(sheet, row, 1, row, 8, {
      font: { name: "Arial", size: 9, color: { argb: COLORS.dark } },
      alignment: { vertical: "middle", wrapText: true },
      border: THIN_BORDER,
    });
    sheet.getCell(row, 2).numFmt = MONEY_FORMAT;
    sheet.getCell(row, 3).numFmt = "#,##0";
    sheet.getCell(row, 4).numFmt = MONEY_FORMAT;
    row += 1;
  }
  const monthlyEndRow = row - 1;
  sheet.mergeCells(row, 1, row, 2);
  sheet.mergeCells(row, 3, row, 4);
  sheet.mergeCells(row, 5, row, 8);
  sheet.getCell(row, 1).value = "Шаг 3. Среднемесячная за квартал";
  sheet.getCell(row, 3).value = formula(`IFERROR(SUM(D${monthlyStartRow}:D${monthlyEndRow})/COUNTIF(B${monthlyStartRow}:B${monthlyEndRow},\">0\"),0)`, numeric(quarterResult.average));
  sheet.getCell(row, 3).numFmt = MONEY_FORMAT;
  sheet.getCell(row, 5).value = `${numeric(quarterResult.monthly_average_sum).toLocaleString("ru-RU")} ÷ ${numeric(quarterResult.months_with_payments)} = ${numeric(quarterResult.average).toLocaleString("ru-RU")} ₽`;
  styleCells(sheet, row, 1, row, 8, {
    fill: { type: "pattern", pattern: "solid", fgColor: { argb: COLORS.light } },
    font: { name: "Arial", size: 9, bold: true, color: { argb: COLORS.dark } },
    alignment: { vertical: "middle", wrapText: true },
    border: THIN_BORDER,
  });
  sheet.getRow(row).height = 28;
  row += 1;

  const verdict = comparison.verdict;
  const verdictColor = verdict === "PASS" ? COLORS.successText : verdict === "FAIL" ? COLORS.dangerText : COLORS.muted;
  const verdictFill = verdict === "PASS" ? COLORS.success : verdict === "FAIL" ? COLORS.danger : COLORS.light;
  sheet.mergeCells(row, 1, row, 2);
  sheet.mergeCells(row, 3, row, 4);
  sheet.mergeCells(row, 5, row, 8);
  sheet.getCell(row, 1).value = "Шаг 4. Сравнение с планкой";
  sheet.getCell(row, 3).value = formula(`C${row - 1}`, numeric(quarterResult.average));
  sheet.getCell(row, 3).numFmt = MONEY_FORMAT;
  sheet.getCell(row, 5).value = `${VERDICT_LABELS[verdict] ?? "Результат неизвестен"}: ${displayNumber(quarterResult.average)} ₽ и ${displayNumber(thresholdResult.used)} ₽`;
  styleCells(sheet, row, 1, row, 8, {
    fill: { type: "pattern", pattern: "solid", fgColor: { argb: verdictFill } },
    font: { name: "Arial", size: 9, bold: true, color: { argb: verdictColor } },
    alignment: { vertical: "middle", wrapText: true },
    border: THIN_BORDER,
  });
  sheet.getRow(row).height = 30;
  row += 2;

  row = addSection(sheet, row, "5. Результат");
  row = addInfoRow(sheet, row, "Применённая планка", toNumber(thresholdResult.used), MONEY_FORMAT);
  row = addInfoRow(sheet, row, "Источник планки", THRESHOLD_LABELS[thresholdResult.source] ?? thresholdResult.source);
  row = addInfoRow(sheet, row, "Среднемесячная выплата за квартал", toNumber(quarterResult.average), MONEY_FORMAT);
  row = addInfoRow(sheet, row, "Сумма месячных средних", toNumber(quarterResult.monthly_average_sum), MONEY_FORMAT);
  row = addInfoRow(sheet, row, "Месяцев с выплатами", toNumber(quarterResult.months_with_payments), "#,##0");
  row = addInfoRow(sheet, row, "Разница с планкой", toNumber(comparison.delta), MONEY_FORMAT);
  row = addInfoRow(sheet, row, "Разница, %", toNumber(comparison.delta_percent), '0.00"%"');
  row = addInfoRow(sheet, row, "Итог", VERDICT_LABELS[verdict] ?? verdict);
  row = addInfoRow(sheet, row, "Рекомендация", RECOMMENDATION_LABELS[recommendation.type] ?? recommendation.type);
  row = addInfoRow(sheet, row, "Изменение на сотрудника в месяц", toNumber(recommendation.salary_change_per_employee_per_month), MONEY_FORMAT);

  sheet.getCell(row - 3, 3).fill = { type: "pattern", pattern: "solid", fgColor: { argb: verdictFill } };
  sheet.getCell(row - 3, 3).font = { name: "Arial", size: 10, bold: true, color: { argb: verdictColor } };
  sheet.pageSetup.printArea = `A1:H${row - 1}`;
  sheet.headerFooter.oddFooter = "&LWageCalc&CОтчёт по выплатам&RСтраница &P из &N";
  return sheet;
}

export async function createCalculationWorkbook(calculation, monthNames){
  if (!calculation?.input || !calculation?.result){
    throw new Error("Для экспорта нужен полный расчёт с исходными данными и результатом");
  }

  const { default: ExcelJS } = await import("exceljs/dist/exceljs.min.js");
  const workbook = new ExcelJS.Workbook();
  workbook.creator = "WageCalc";
  workbook.created = new Date();
  workbook.modified = new Date();
  workbook.calcProperties.fullCalcOnLoad = true;
  buildReportSheet(workbook, calculation, monthNames);
  return workbook;
}

export async function exportCalculationReport(calculation, monthNames){
  const workbook = await createCalculationWorkbook(calculation, monthNames);
  const buffer = await workbook.xlsx.writeBuffer();
  const blob = new Blob([buffer], {
    type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  const period = calculation.input.period ?? {};
  link.href = url;
  link.download = `${sanitizeFileName(calculation.input.title)}_${period.year ?? ""}_Q${period.quarter ?? ""}.xlsx`;
  document.body.appendChild(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 0);
}
