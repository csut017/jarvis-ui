import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { tap, catchError, map } from 'rxjs/operators';
import { List } from './common';
import { Result, Results } from './result';

@Injectable({
  providedIn: 'root'
})
export class SourceService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('SourceService');
  }

  private logger: Logger;

  list(): Observable<Results<Source>> {
    const url = `${environment.apiURL}sources`;
    this.logger.log(`Retrieving sources`, url);
    return this.http.get<List<Source>>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved sources', res)),
        map(res => Results.new(res ? res.items : [])),
        catchError(this.logger.handleError(`list()`, Results.new<Source>(null, 'Unable to retrieve sources')))
      );
  }

  get(name: string): Observable<Result<Source>> {
    const uriName = encodeURIComponent(name);
    const url = `${environment.apiURL}sources/${uriName}`;
    this.logger.log(`Retrieving source details for ${name}`, url);
    return this.http.get<Source>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved source details', res)),
        map(res => Result.new(res)),
        catchError(this.logger.handleError(`get('${name}')`, Result.new<Source>(null, 'Unable to retrieve source'))),
      );
  }
}

export interface Source {
  name: string;
  status: string;
  sensors: string[];
  effectors: string[];
}
