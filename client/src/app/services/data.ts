export interface ValueList {
    station: string;
    count: number,
    items: SourceItem[]
}

export interface SourceItem {
    source: string,
    time: string,
    count: number,
    values: SourceValue[]
}

export interface SourceValue {
    name: string,
    value: number
}