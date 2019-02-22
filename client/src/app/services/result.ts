export class Result<T> {
    item: T;
    success: boolean;
    message: string;

    static new<T>(item: T, message?: string): Result<T> {
        let out = new Result<T>();
        out.item = item;
        out.success = !message;
        out.message = message;
        return out;
    }
}

export class Results<T> {
    items: T[];
    success: boolean;
    message: string;

    static new<T>(items: T[], message?: string): Results<T> {
        let out = new Results<T>();
        out.items = items;
        out.success = !message;
        out.message = message;
        return out;
    }
}

export interface Status {
    status: string;
    msg: string;
}
