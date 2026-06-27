// NOTE: Google Apps Script does not run TypeScript natively.
// Compile this file to JS before pushing: tsgo --project tsconfig.json
// The compiled output (update-beggers-sheet.js) is what gets deployed to Apps Script.

declare const SpreadsheetApp: any;
declare const ContentService: any;

const SPREADSHEET_ID = "1HHoFW44z9gD1J0MfqBKAh5WkTFYIlMDUH7Fks3LK2bM";
const DEFAULT_PAGE_SIZE = 10;
const COLUMNS = ["companyName", "link", "dateApplied", "industry", "phoneNumber", "email", "status", "notes"] as const;
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

type Column = typeof COLUMNS[number];
type Action = "create" | "patch" | "delete";

interface PostData {
  length: number;
  type: string;
  contents: string;
  name: string;
}

interface WebAppEvent {
  queryString: string;
  parameter: Record<string, string>;
  parameters: Record<string, string[]>;
  pathInfo: string;
  contextPath: string;
  contentLength: number;
  postData: PostData;
}

interface Entry {
  companyName: string;
  link: string;
  dateApplied: string | Date;
  industry: string;
  phoneNumber: string;
  email: string;
  status: string;
  notes: string;
}

interface PatchRequest {
  action: Action;
  matchBy: Partial<Entry>;
  update: Partial<Entry>;
}

interface DeleteRequest {
  action: Action;
  matchBy: Partial<Entry>;
}

interface QueryFilters {
  page: number;
  pageSize: number;
  search: string;
  industry: string;
  status: string;
  order: "asc" | "desc";
}

interface ApiResponse {
  status: "success" | "error";
  message?: string;
  messages?: string[];
  rows?: Entry[];
  page?: number;
  pageSize?: number;
  totalPages?: number;
  totalRows?: number;
}

type TextOutput = any;

function jsonOutput(obj: ApiResponse): TextOutput {
  return ContentService
    .createTextOutput(JSON.stringify(obj))
    .setMimeType(ContentService.MimeType.JSON);
}

function doPost(e: WebAppEvent): TextOutput {
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
    let result: string | null = null;

    switch (action) {
      case "create": {
        const entry: Entry = body;
        const errors = validate(entry, action);
        if (errors.length > 0) return jsonOutput({ status: "error", messages: errors });
        result = beggersSheet.create(entry);
        break;
      }
      case "patch": {
        const req: PatchRequest = body;
        const errors = validate(req, action);
        if (errors.length > 0) return jsonOutput({ status: "error", messages: errors });
        result = beggersSheet.patch(req);
        break;
      }
      case "delete": {
        const req: DeleteRequest = body;
        const errors = validate(req, action);
        if (errors.length > 0) return jsonOutput({ status: "error", messages: errors });
        result = beggersSheet.delete(req);
        break;
      }
    }

    if (result) return jsonOutput({ status: "error", message: result });
    return jsonOutput({ status: "success" });
  } catch (err: any) {
    return jsonOutput({ status: "error", message: err.message });
  }
}

function doGet(e: WebAppEvent): TextOutput {
  try {
    const params = e.parameter || {};
    const order = params.order === "asc" ? "asc" : "desc";
    const filters: QueryFilters = {
      page: parseInt(params.page) || 1,
      pageSize: parseInt(params.pageSize) || DEFAULT_PAGE_SIZE,
      search: params.search || "",
      industry: params.industry || "",
      status: params.status || "",
      order
    };

    const beggersSheet = new Sheet(SPREADSHEET_ID);
    return jsonOutput(beggersSheet.read(filters));
  } catch (err: any) {
    return jsonOutput({ status: "error", message: err.message });
  }
}

function validate(entry: Entry | PatchRequest | DeleteRequest, action: Action): string[] {
  const errors: string[] = [];

  if (action === "create") {
    const e = entry as Entry;
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
  } else if (action === "patch") {
    const req = entry as PatchRequest;
    const hasMatch = COLUMNS.some(col => (req.matchBy as any)[col] != null && String((req.matchBy as any)[col]).trim() !== "");
    if (!hasMatch) errors.push("matchBy requires at least one field");
    const hasUpdate = COLUMNS.some(col => (req.update as any)[col] != null && String((req.update as any)[col]).trim() !== "");
    if (!hasUpdate) errors.push("update requires at least one field");
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
  } else {
    const req = entry as DeleteRequest;
    const hasMatch = COLUMNS.some(col => (req.matchBy as any)[col] != null && String((req.matchBy as any)[col]).trim() !== "");
    if (!hasMatch) errors.push("matchBy requires at least one field");
  }

  return errors;
}

class Sheet {
  private ss: any;
  private sheet: any;

  constructor(spreadsheetId: string) {
    this.ss = SpreadsheetApp.openById(spreadsheetId);
    this.sheet = this.ss.getSheets()[0];
  }

  private _findRow(filters: Partial<Entry>): number {
    const values = this.sheet.getDataRange().getValues();
    const matchCols = COLUMNS.filter(col => (filters as any)[col] != null);
    for (let i = 1; i < values.length; i++) {
      const isMatch = matchCols.every(col => {
        const colIdx = COLUMNS.indexOf(col);
        const cellValue = values[i][colIdx];
        const filterValue = (filters as any)[col];
        const cellDate = new Date(cellValue).getTime();
        const filterDate = new Date(filterValue).getTime();
        if (!isNaN(cellDate) && !isNaN(filterDate)) {
          return cellDate === filterDate;
        }
        return String(cellValue) === String(filterValue);
      });
      if (isMatch) return i + 1;
    }
    return -1;
  }

  private _rowToObject(row: any[]): Entry {
    return Object.fromEntries(COLUMNS.map((col, i) => [col, row[i]])) as Entry;
  }

  create(entry: Entry): string | null {
    const row = COLUMNS.map(col => (entry as any)[col] != null ? (entry as any)[col] : "");
    this.sheet.appendRow(row);
    return null;
  }

  patch(req: PatchRequest): string | null {
    const rowNumber = this._findRow(req.matchBy);
    if (rowNumber === -1) {
      return `No entry found matching: ${JSON.stringify(req.matchBy)}`;
    }

    // update row
    const currentRow = this.sheet.getRange(rowNumber, 1, 1, COLUMNS.length).getValues()[0];
    const updatedRow = COLUMNS.map((col, i) =>
      (req.update as any)[col] != null ? (req.update as any)[col] : currentRow[i]
    );

    const rows = this.sheet.getRange(rowNumber, 1, 1, COLUMNS.length);
    rows.setValues([updatedRow]);

    return null;
  }

  delete(req: DeleteRequest): string | null {
    const rowNumber = this._findRow(req.matchBy);
    if (rowNumber === -1) {
      return `No entry found matching: ${JSON.stringify(req.matchBy)}`;
    }

    this.sheet.deleteRow(rowNumber);
    return null;
  }

  read(filters: QueryFilters): ApiResponse {
    const data = this.sheet.getDataRange().getValues();

    if (data.length <= 1) {
      return { status: "success", rows: [], page: filters.page, totalPages: 0, totalRows: 0 };
    }

    let rows: Entry[] = data.slice(1).map((row: any[]) => this._rowToObject(row));


    // dateApplied sort, desc by default
    rows.sort((a, b) => {
      const da = new Date(a.dateApplied as any).getTime() || 0;
      const db = new Date(b.dateApplied as any).getTime() || 0;
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
      rows = rows.filter(r =>
        (r.companyName && String(r.companyName).toLowerCase().includes(q)) ||
        (r.link && String(r.link).toLowerCase().includes(q)) ||
        (r.email && String(r.email).toLowerCase().includes(q)) ||
        (r.notes && String(r.notes).toLowerCase().includes(q))
      );
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
