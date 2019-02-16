import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class LoggingService {

  constructor() { }

  get(source: string) {
    return new Logger(source);
  }
}

export class Logger {
  constructor(private source: string) { }

  log(msg: string, data?: any, includeTrace?: boolean) {
    if (data || includeTrace) {
      console.groupCollapsed(`[${this.source}] ${msg}`);
      if (data) console.log(data);
      if (includeTrace) console.trace();
      console.groupEnd();
    } else {
      console.log(`[${this.source}] ${msg}`);
    }
  }

  handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.groupCollapsed(`[${this.source}] ${operation} failed: ${error.message}`);
      console.error(error);
      console.trace();
      console.groupEnd();
      return of(result as T);
    };
  }
}