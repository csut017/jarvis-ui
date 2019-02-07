import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { tap, catchError, map } from 'rxjs/operators';
import { List } from './common';

@Injectable({
  providedIn: 'root'
})
export class SourceService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('SourceService');
  }

  private logger: Logger;

  list(): Observable<Source[]> {
    const url = `${environment.apiURL}sources`;
    this.logger.log(`Retrieving sources`, url);
    return this.http.get<List<Source>>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved sources', res)),
        catchError(this.logger.handleError<List<Source>>(`list()`)),
        map(res => res ? res.items : [])
      );
  }

  get(name: string): Observable<Source> {
    const uriName = encodeURIComponent(name);
    const url = `${environment.apiURL}stations/${uriName}`;
    this.logger.log(`Retrieving source details for ${name}`, url);
    return this.http.get<Source>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved source details', res)),
        catchError(this.logger.handleError<Source>(`get('${name}')`)),
      );
  }
}

export interface Source {
  name: string;
  status: string;
  sensors: string[];
  effectors: string[];
}
