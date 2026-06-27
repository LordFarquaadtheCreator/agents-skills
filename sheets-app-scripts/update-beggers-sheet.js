"use strict";
// NOTE: Google Apps Script does not run TypeScript natively.
// Compile this file to JS before pushing: tsgo --project tsconfig.json
// The compiled output (update-beggers-sheet.js) is what gets deployed to Apps Script.
const SPREADSHEET_ID = "1HHoFW44z9gD1J0MfqBKAh5WkTFYIlMDUH7Fks3LK2bM";
const DEFAULT_PAGE_SIZE = 10;
const COLUMNS = ["companyName", "link", "dateApplied", "industry", "phoneNumber", "email", "status", "notes"];
const VALID_ACTIONS = new Set(["create", "patch", "delete"]);
const VALID_STATUSES = new Set([
    "Not Started",
    "Applied Only",
    "Applied + Emailed",
    "Applied + Called",
    "Applied + Emailed + Called",
    "Interview!",
    "Got the Job!",
    "Didn't Get It"
]);
const VALID_INDUSTRIES = new Set([
    "Tech",
    "Health Care",
    "Retail",
    "Finance",
    "Gig",
    "Other"
]);
function jsonOutput(obj) {
    return ContentService
        .createTextOutput(JSON.stringify(obj))
        .setMimeType(ContentService.MimeType.JSON);
}
function doPost(e) {
    try {
        if (!e.postData || e.postData.type !== "application/json") {
            return jsonOutput({ status: "error", message: "Content-Type must be application/json" });
        }
        const body = JSON.parse(e.postData.contents);
        const action = body.action;
        if (!action || !VALID_ACTIONS.has(action)) {
            return jsonOutput({ status: "error", message: `Invalid action "${action}". Must be one of: ${[...VALID_ACTIONS].join(", ")}` });
        }
        const beggersSheet = new Sheet(SPREADSHEET_ID);
        let result = null;
        switch (action) {
            case "create": {
                const entry = body;
                const errors = validate(entry, action);
                if (errors.length > 0)
                    return jsonOutput({ status: "error", messages: errors });
                result = beggersSheet.create(entry);
                break;
            }
            case "patch": {
                const req = body;
                const errors = validate(req, action);
                if (errors.length > 0)
                    return jsonOutput({ status: "error", messages: errors });
                result = beggersSheet.patch(req);
                break;
            }
            case "delete": {
                const req = body;
                const errors = validate(req, action);
                if (errors.length > 0)
                    return jsonOutput({ status: "error", messages: errors });
                result = beggersSheet.delete(req);
                break;
            }
        }
        if (result)
            return jsonOutput({ status: "error", message: result });
        return jsonOutput({ status: "success" });
    }
    catch (err) {
        return jsonOutput({ status: "error", message: err.message });
    }
}
function doGet(e) {
    try {
        const params = e.parameter || {};
        const order = params.order === "asc" ? "asc" : "desc";
        const filters = {
            page: parseInt(params.page) || 1,
            pageSize: parseInt(params.pageSize) || DEFAULT_PAGE_SIZE,
            search: params.search || "",
            industry: params.industry || "",
            status: params.status || "",
            order
        };
        const beggersSheet = new Sheet(SPREADSHEET_ID);
        return jsonOutput(beggersSheet.read(filters));
    }
    catch (err) {
        return jsonOutput({ status: "error", message: err.message });
    }
}
function validate(entry, action) {
    const errors = [];
    if (action === "create") {
        const e = entry;
        if (!e.status || !VALID_STATUSES.has(e.status)) {
            errors.push(`Invalid status "${e.status}". Must be one of: ${[...VALID_STATUSES].join(", ")}`);
        }
        if (!e.industry || !VALID_INDUSTRIES.has(e.industry)) {
            errors.push(`Invalid industry "${e.industry}". Must be one of: ${[...VALID_INDUSTRIES].join(", ")}`);
        }
        if (e.phoneNumber != null) {
            const cleaned = e.phoneNumber.replace(/[\s\-().+]/g, "");
            if (!/^\d{10,15}$/.test(cleaned)) {
                errors.push(`Invalid phone number "${e.phoneNumber}". Must contain 10-15 digits.`);
            }
        }
    }
    else if (action === "patch") {
        const req = entry;
        const hasMatch = COLUMNS.some(col => req.matchBy[col] != null && String(req.matchBy[col]).trim() !== "");
        if (!hasMatch)
            errors.push("matchBy requires at least one field");
        const hasUpdate = COLUMNS.some(col => req.update[col] != null && String(req.update[col]).trim() !== "");
        if (!hasUpdate)
            errors.push("update requires at least one field");
        if (req.update.status != null && !VALID_STATUSES.has(req.update.status)) {
            errors.push(`Invalid status "${req.update.status}". Must be one of: ${[...VALID_STATUSES].join(", ")}`);
        }
        if (req.update.industry != null && !VALID_INDUSTRIES.has(req.update.industry)) {
            errors.push(`Invalid industry "${req.update.industry}". Must be one of: ${[...VALID_INDUSTRIES].join(", ")}`);
        }
        if (req.update.phoneNumber != null) {
            const cleaned = req.update.phoneNumber.replace(/[\s\-().+]/g, "");
            if (!/^\d{10,15}$/.test(cleaned)) {
                errors.push(`Invalid phone number "${req.update.phoneNumber}". Must contain 10-15 digits.`);
            }
        }
    }
    else {
        const req = entry;
        const hasMatch = COLUMNS.some(col => req.matchBy[col] != null && String(req.matchBy[col]).trim() !== "");
        if (!hasMatch)
            errors.push("matchBy requires at least one field");
    }
    return errors;
}
class Sheet {
    constructor(spreadsheetId) {
        this.ss = SpreadsheetApp.openById(spreadsheetId);
        this.sheet = this.ss.getSheets()[0];
    }
    _findRow(filters) {
        const values = this.sheet.getDataRange().getValues();
        const matchCols = COLUMNS.filter(col => filters[col] != null);
        for (let i = 1; i < values.length; i++) {
            const isMatch = matchCols.every(col => {
                const colIdx = COLUMNS.indexOf(col);
                const cellValue = values[i][colIdx];
                const filterValue = filters[col];
                const cellDate = new Date(cellValue).getTime();
                const filterDate = new Date(filterValue).getTime();
                if (!isNaN(cellDate) && !isNaN(filterDate)) {
                    return cellDate === filterDate;
                }
                return String(cellValue) === String(filterValue);
            });
            if (isMatch)
                return i + 1;
        }
        return -1;
    }
    _rowToObject(row) {
        return Object.fromEntries(COLUMNS.map((col, i) => [col, row[i]]));
    }
    create(entry) {
        const row = COLUMNS.map(col => entry[col] != null ? entry[col] : "");
        this.sheet.appendRow(row);
        return null;
    }
    patch(req) {
        const rowNumber = this._findRow(req.matchBy);
        if (rowNumber === -1) {
            return `No entry found matching: ${JSON.stringify(req.matchBy)}`;
        }
        // update row
        const currentRow = this.sheet.getRange(rowNumber, 1, 1, COLUMNS.length).getValues()[0];
        const updatedRow = COLUMNS.map((col, i) => req.update[col] != null ? req.update[col] : currentRow[i]);
        const rows = this.sheet.getRange(rowNumber, 1, 1, COLUMNS.length);
        rows.setValues([updatedRow]);
        return null;
    }
    delete(req) {
        const rowNumber = this._findRow(req.matchBy);
        if (rowNumber === -1) {
            return `No entry found matching: ${JSON.stringify(req.matchBy)}`;
        }
        this.sheet.deleteRow(rowNumber);
        return null;
    }
    read(filters) {
        const data = this.sheet.getDataRange().getValues();
        if (data.length <= 1) {
            return { status: "success", rows: [], page: filters.page, totalPages: 0, totalRows: 0 };
        }
        let rows = data.slice(1).map((row) => this._rowToObject(row));
        // dateApplied sort, desc by default
        rows.sort((a, b) => {
            const da = new Date(a.dateApplied).getTime() || 0;
            const db = new Date(b.dateApplied).getTime() || 0;
            return filters.order === "asc" ? da - db : db - da;
        });
        if (filters.industry) {
            rows = rows.filter(r => r.industry === filters.industry);
        }
        if (filters.status) {
            rows = rows.filter(r => r.status === filters.status);
        }
        if (filters.search) {
            const q = filters.search.toLowerCase();
            rows = rows.filter(r => (r.companyName && String(r.companyName).toLowerCase().includes(q)) ||
                (r.link && String(r.link).toLowerCase().includes(q)) ||
                (r.email && String(r.email).toLowerCase().includes(q)) ||
                (r.notes && String(r.notes).toLowerCase().includes(q)));
        }
        const totalRows = rows.length;
        const totalPages = Math.ceil(totalRows / filters.pageSize);
        const start = (filters.page - 1) * filters.pageSize;
        const paged = rows.slice(start, start + filters.pageSize);
        return {
            status: "success",
            rows: paged,
            page: filters.page,
            pageSize: filters.pageSize,
            totalPages,
            totalRows
        };
    }
}
