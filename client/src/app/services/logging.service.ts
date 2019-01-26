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

  log(msg: string, data?: any) {
    if (data) {
      console.groupCollapsed(`[${this.source}] ${msg}`);
      console.log(data);
      console.groupEnd();
    } else {
      console.log(`[${this.source}] ${msg}`);
    }
  }

  handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.error(error);
      this.log(`${operation} failed: ${error.message}`);
      return of(result as T);
    };
  }
}